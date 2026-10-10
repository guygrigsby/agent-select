package collector

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// HTTPFetcher is the production Fetcher: descriptive User-Agent, strict
// timeouts, https-only with private-range redirect rejection outside test
// mode, a hard body cap and Retry-After parsing in both forms.
type HTTPFetcher struct {
	Client    *http.Client
	UserAgent string
	MaxBody   int64
	Timeout   time.Duration
	// InsecureAllowLocal permits http and private hosts. Test mode only;
	// the daemon never sets it.
	InsecureAllowLocal bool

	// resolve and dial back guardedDial; tests inject stubs to verify
	// resolution, validation and the dialed address without network I/O.
	resolve func(ctx context.Context, host string) ([]netip.Addr, error)
	dial    func(ctx context.Context, network, addr string) (net.Conn, error)
}

// NewHTTPFetcher builds the production fetcher: HTTPS-only, private
// networks unreachable.
func NewHTTPFetcher(userAgent string) *HTTPFetcher {
	return newHTTPFetcher(userAgent, false)
}

// NewLocalTestFetcher builds a fetcher that permits http:// and private
// addresses. Tests only: never wire into the production daemon.
func NewLocalTestFetcher(userAgent string) *HTTPFetcher {
	return newHTTPFetcher(userAgent, true)
}

func newHTTPFetcher(userAgent string, allowLocal bool) *HTTPFetcher {
	f := &HTTPFetcher{
		UserAgent:          userAgent,
		MaxBody:            10 << 20,
		Timeout:            30 * time.Second,
		InsecureAllowLocal: allowLocal,
		resolve:            resolveHost,
		dial:               (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
	}
	transport := &http.Transport{
		DialContext: f.dial,
	}
	if !f.InsecureAllowLocal {
		transport.DialContext = f.guardedDial
	}
	f.Client = &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return f.checkRedirect(req, via)
		},
		Timeout: f.Timeout,
	}
	return f
}

// guardedDial resolves the host, validates every resolved address and dials
// the validated address itself. Literal IPs skip resolution. Dialing exactly
// the validated address (not the hostname) closes the DNS-rebinding window:
// the transport cannot re-resolve to a different host between check and
// connect. SNI and Host stay derived from the URL hostname.
func (f *HTTPFetcher) guardedDial(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("collector: bad address %s: %w", addr, err)
	}
	addrs, err := f.resolve(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("collector: %s resolved to no addresses", host)
	}
	for _, a := range addrs {
		if isPrivateIP(a) {
			return nil, fmt.Errorf("collector: %s resolves to private address %s", host, a)
		}
	}
	return f.dial(ctx, network, net.JoinHostPort(addrs[0].String(), port))
}

// resolveHost resolves a hostname to its addresses; literal IPs pass through
// unchanged.
func resolveHost(ctx context.Context, host string) ([]netip.Addr, error) {
	if ip, err := netip.ParseAddr(host); err == nil {
		return []netip.Addr{ip.Unmap()}, nil
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("collector: %s resolved to no addresses", host)
	}
	out := make([]netip.Addr, 0, len(ips))
	for _, ip := range ips {
		a, ok := netip.AddrFromSlice(ip.IP)
		if !ok {
			return nil, fmt.Errorf("collector: %s resolved to unparseable address %s", host, ip.IP)
		}
		out = append(out, a.Unmap())
	}
	return out, nil
}

func (f *HTTPFetcher) checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 5 {
		return fmt.Errorf("collector: stopped after 5 redirects")
	}
	if err := f.validateURL(req.URL); err != nil {
		return err
	}
	return nil
}

func (f *HTTPFetcher) validateURL(u *url.URL) error {
	if f.InsecureAllowLocal {
		if u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf("collector: scheme %s not allowed", u.Scheme)
		}
		return nil
	}
	if u.Scheme != "https" {
		return fmt.Errorf("collector: scheme %s not allowed", u.Scheme)
	}
	// Literal IP hosts are checked here; hostnames are resolved and
	// checked at dial time (guardedDial).
	if ip, err := netip.ParseAddr(u.Hostname()); err == nil && isPrivateIP(ip.Unmap()) {
		return fmt.Errorf("collector: refusing private host %s", u.Hostname())
	}
	return nil
}

// Fetch performs one GET. Non-2xx statuses are returned as results (the
// worker classifies); transport errors as errors.
func (f *HTTPFetcher) Fetch(ctx context.Context, sourceID, resourceURL string) (FetchResult, error) {
	u, err := url.Parse(resourceURL)
	if err != nil {
		return FetchResult{}, fmt.Errorf("collector: bad resource url: %w", err)
	}
	if err := f.validateURL(u); err != nil {
		return FetchResult{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resourceURL, nil)
	if err != nil {
		return FetchResult{}, err
	}
	ua := f.UserAgent
	if ua == "" {
		ua = "agent-routing-collector/0.1 (+https://github.com/guygrigsby/agent-select)"
	}
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "application/json, text/html;q=0.9, */*;q=0.5")
	resp, err := f.Client.Do(req)
	if err != nil {
		return FetchResult{}, err
	}
	defer resp.Body.Close()
	if resp.ContentLength > f.MaxBody {
		return FetchResult{HTTPStatus: resp.StatusCode}, fmt.Errorf(
			"collector: body of %d bytes exceeds cap %d", resp.ContentLength, f.MaxBody)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, f.MaxBody+1))
	if err != nil {
		return FetchResult{}, err
	}
	if int64(len(body)) > f.MaxBody {
		return FetchResult{HTTPStatus: resp.StatusCode}, fmt.Errorf(
			"collector: body exceeds cap %d", f.MaxBody)
	}
	headers := canonicalHeaderBytes(resp)
	return FetchResult{
		HTTPStatus: resp.StatusCode,
		Body:       body,
		Headers:    headers,
		RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After"), time.Now()),
	}, nil
}

// canonicalHeaderBytes renders selected response headers as canonical
// bytes for the evidence envelope.
func canonicalHeaderBytes(resp *http.Response) []byte {
	var b strings.Builder
	for _, name := range []string{"etag", "last-modified", "cache-control", "content-type"} {
		if v := resp.Header.Get(name); v != "" {
			fmt.Fprintf(&b, "%s: %s\n", name, v)
		}
	}
	return []byte(b.String())
}

// parseRetryAfter honors both the seconds and HTTP-date forms.
func parseRetryAfter(v string, now time.Time) time.Duration {
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil && secs >= 0 {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil && t.After(now) {
		return t.Sub(now)
	}
	return 0
}
