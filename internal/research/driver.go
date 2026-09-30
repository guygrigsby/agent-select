package research

import (
	"runtime"
	"runtime/debug"
	"strings"
)

// linkedDriverInfo reports the modernc.org/sqlite module version linked into
// this binary (empty when unavailable, e.g. under test without build info).
func linkedDriverInfo() (module, version string) {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "", ""
	}
	for _, d := range bi.Deps {
		if d.Path == "modernc.org/sqlite" {
			return d.Path, d.Version
		}
	}
	return "", ""
}

// toolchainMinor renders runtime.Version() as a pinned minor like "go1.26".
func toolchainMinor() string {
	v := runtime.Version() // e.g. go1.26.2
	parts := strings.SplitN(strings.TrimPrefix(v, "go"), ".", 3)
	if len(parts) < 2 {
		return v
	}
	return "go" + parts[0] + "." + parts[1]
}

// CheckDriverPin compares the linked driver and toolchain against build.json
// pins (§2.2); mismatch is exit 3. Placeholders never mismatch.
func CheckDriverPin(cfg *Config) error {
	if cfg.Driver.IsPlaceholder() {
		return nil
	}
	mod, ver := linkedDriverInfo()
	if mod == "" {
		return inconsf("driver pin: linked modernc.org/sqlite version unavailable from build info")
	}
	if mod != cfg.Driver.SQLiteModule {
		return inconsf("driver pin: build.json pins module %s but binary links %s", cfg.Driver.SQLiteModule, mod)
	}
	if ver != cfg.Driver.SQLiteVersion {
		return inconsf("driver pin: build.json pins %s %s but binary links %s %s (byte reproducibility requires the pinned driver)", mod, cfg.Driver.SQLiteVersion, mod, ver)
	}
	if tc := toolchainMinor(); tc != cfg.Driver.GoToolchain {
		return inconsf("driver pin: build.json pins toolchain %s but binary was built with %s", cfg.Driver.GoToolchain, tc)
	}
	return nil
}

// DriverSummary describes the running tool for --json output.
func DriverSummary() map[string]any {
	mod, ver := linkedDriverInfo()
	return map[string]any{
		"sqlite_driver_module":  mod,
		"sqlite_driver_version": ver,
		"go_toolchain":          runtime.Version(),
		"go_toolchain_minor":    toolchainMinor(),
	}
}
