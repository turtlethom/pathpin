package cmd

import (
	"bytes"
	"runtime/debug"
	"testing"

	"github.com/turtlethom/pathpin/internal/shell"
)

// buildInfo fakes debug.ReadBuildInfo with this main module version.
func buildInfo(v string) func() (*debug.BuildInfo, bool) {
	return func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: v}}, true
	}
}

// noBuildInfo fakes a binary built without module support.
func noBuildInfo() (*debug.BuildInfo, bool) { return nil, false }

func TestVersionString(t *testing.T) {
	for _, tc := range []struct {
		name, stamped string
		info          func() (*debug.BuildInfo, bool)
		want          string
	}{
		{"GoReleaser's stamp wins", "2.0.0", buildInfo("v1.9.0"), "2.0.0"},
		{"a stamped v is dropped", "v2.0.0", noBuildInfo, "2.0.0"},
		{"go install @v2.0.0", "", buildInfo("v2.0.0"), "2.0.0"},
		{"local go build", "", buildInfo("v0.0.0-20261025174224-32dc54d89a2d+dirty"),
			"0.0.0-20261025174224-32dc54d89a2d+dirty"},
		{"go run or a build outside Git", "", buildInfo("(devel)"), "dev"},
		{"empty build info", "", buildInfo(""), "dev"},
		{"no build info", "", noBuildInfo, "dev"},
	} {
		if got := versionString(tc.stamped, tc.info); got != tc.want {
			t.Errorf("%s: versionString(%q) = %q, want %q", tc.name, tc.stamped, got, tc.want)
		}
	}
}

func TestVersionFlag(t *testing.T) {
	e := newRootEnv(t)
	// A config that doesn't parse: --version must not load it.
	writeFile(t, e.configPath, "not toml [[[")
	t.Setenv(shell.CmdEnv, "pp") // the binary's name even with --cmd pp

	for _, flag := range []string{"--version", "-v"} {
		root := NewRootCmd("2.0.0") // what Execute builds; the menu never opens
		var out, errOut bytes.Buffer
		root.SetOut(&out)
		root.SetErr(&errOut)
		root.SetArgs([]string{flag})
		if err := root.Execute(); err != nil {
			t.Fatalf("pathpin %s: %v", flag, err)
		}
		if got, want := out.String(), "pathpin 2.0.0\n"; got != want {
			t.Errorf("pathpin %s printed %q, want %q", flag, got, want)
		}
		if errOut.Len() != 0 {
			t.Errorf("pathpin %s wrote to stderr: %q", flag, errOut.String())
		}
	}
	if e.cdTarget(t) != "" {
		t.Error("--version must not cd")
	}
}
