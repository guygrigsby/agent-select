package routing

import (
	"strings"
	"sync"
	"testing"
)

func TestNewEngineNilUnavailable(t *testing.T) {
	e := NewEngine(nil)
	if e.ActiveHash() != "" {
		t.Fatal("nil engine has a hash")
	}
	if e.pinned() != nil {
		t.Fatal("nil engine pinned a snapshot")
	}
	// A zero-value Snapshot pointer (hash empty) is also unavailable.
	var zero Snapshot
	if NewEngine(&zero).ActiveHash() != "" {
		t.Fatal("zero-value snapshot accepted")
	}
}

func TestAdoptSwapsGeneration(t *testing.T) {
	e := NewEngine(mustSnapshot(t))
	first := e.ActiveHash()
	if first == "" {
		t.Fatal("engine unavailable")
	}
	// Build a second valid snapshot: single offer variant.
	d := fixtureSnapshotData()
	d.Offers = d.Offers[:1]
	s2, err := NewSnapshot(d)
	if err != nil {
		t.Fatal(err)
	}
	enc2, err := EncodeSnapshot(s2)
	if err != nil {
		t.Fatal(err)
	}
	if err := e.Adopt(enc2); err != nil {
		t.Fatalf("adopt failed: %v", err)
	}
	if e.ActiveHash() != s2.Hash() || e.ActiveHash() == first {
		t.Fatal("generation not swapped")
	}
}

func TestFailedAdoptionRetainsPriorGeneration(t *testing.T) {
	e := NewEngine(mustSnapshot(t))
	prior := e.ActiveHash()
	bad := []string{
		"",
		"nonsense",
		`{"schema":2}`,
		fixtureCanonicalBytes[:20],
		`{}`,
	}
	for _, b := range bad {
		if err := e.Adopt([]byte(b)); err == nil {
			t.Fatalf("adopt accepted %q", b)
		}
		if e.ActiveHash() != prior {
			t.Fatalf("failed adoption changed generation (input %q)", b)
		}
	}
	// Corrupted valid bytes: each corruption either fails (retaining the
	// prior generation) or decodes to different-but-self-consistent
	// canonical bytes (a legitimate swap); self-consistency is covered by
	// TestCorruptBytesRejected. Here we assert retention for every failing
	// corruption.
	enc, err := EncodeSnapshot(mustSnapshot(t))
	if err != nil {
		t.Fatal(err)
	}
	for i := 10; i < len(enc); i += 53 {
		corrupt := append([]byte(nil), enc...)
		corrupt[i] ^= 0x01
		if err := e.Adopt(corrupt); err == nil {
			prior = e.ActiveHash() // legitimate self-consistent swap
			continue
		}
		if e.ActiveHash() != prior {
			t.Fatalf("failed adoption changed generation at %d", i)
		}
	}
}

func TestAdoptUnavailableEngineBecomesAvailable(t *testing.T) {
	e := NewEngine(nil)
	enc, err := EncodeSnapshot(mustSnapshot(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := e.Adopt(enc); err != nil {
		t.Fatal(err)
	}
	if e.ActiveHash() == "" {
		t.Fatal("engine still unavailable after valid adopt")
	}
}

// TestAdoptRace hammers Adopt from multiple writers while readers observe
// ActiveHash and pinned data. Run with -race. Invariant: after the first
// successful adoption, every reader sees a nonempty hash belonging to one of
// the two valid generations.
func TestAdoptRace(t *testing.T) {
	d1 := fixtureSnapshotData()
	s1, err := NewSnapshot(d1)
	if err != nil {
		t.Fatal(err)
	}
	d2 := fixtureSnapshotData()
	d2.Offers = d2.Offers[:1]
	s2, err := NewSnapshot(d2)
	if err != nil {
		t.Fatal(err)
	}
	e1, err := EncodeSnapshot(s1)
	if err != nil {
		t.Fatal(err)
	}
	e2, err := EncodeSnapshot(s2)
	if err != nil {
		t.Fatal(err)
	}
	e := NewEngine(s1)
	valid := map[string]bool{s1.Hash(): true, s2.Hash(): true}

	var wg sync.WaitGroup
	stop := make(chan struct{})
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; ; i++ {
				select {
				case <-stop:
					return
				default:
				}
				b := e1
				if i%2 == 0 {
					b = e2
				}
				if err := e.Adopt(b); err != nil {
					t.Errorf("valid adopt failed: %v", err)
					return
				}
			}
		}()
	}
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				h := e.ActiveHash()
				if h == "" || !valid[h] {
					t.Errorf("reader saw invalid hash %q", h)
					return
				}
				if p := e.pinned(); p == nil || !valid[p.Hash()] {
					t.Errorf("reader pinned invalid generation")
					return
				}
				_ = e.pinned().Data()
			}
		}()
	}
	// Let the goroutines interleave; bounded work, then stop.
	done := make(chan struct{})
	go func() {
		for i := range 20000 {
			_ = i
		}
		close(done)
	}()
	<-done
	close(stop)
	wg.Wait()
}

func TestErrorShape(t *testing.T) {
	_, err := DecodeSnapshot([]byte("nonsense"))
	e, ok := err.(*Error)
	if !ok {
		t.Fatalf("error type %T", err)
	}
	if e.Code != CodeInvalidSnapshot {
		t.Fatalf("code %s", e.Code)
	}
	if !strings.Contains(e.Error(), CodeInvalidSnapshot) {
		t.Fatal("message lacks code")
	}
}
