package oauthflow

import (
	"reflect"
	"testing"
)

func TestBrowserCommand(t *testing.T) {
	cases := []struct {
		goos     string
		wantName string
		wantArgs []string
	}{
		{"darwin", "open", []string{"https://example.com/x"}},
		{"linux", "xdg-open", []string{"https://example.com/x"}},
		{"windows", "rundll32", []string{"url.dll,FileProtocolHandler", "https://example.com/x"}},
	}

	for _, c := range cases {
		t.Run(c.goos, func(t *testing.T) {
			name, args, err := browserCommand(c.goos, "https://example.com/x")
			if err != nil {
				t.Fatalf("browserCommand(%q, ...) returned error: %v", c.goos, err)
			}
			if name != c.wantName {
				t.Errorf("name = %q, want %q", name, c.wantName)
			}
			if !reflect.DeepEqual(args, c.wantArgs) {
				t.Errorf("args = %v, want %v", args, c.wantArgs)
			}
		})
	}
}

func TestBrowserCommand_UnsupportedOS(t *testing.T) {
	_, _, err := browserCommand("plan9", "https://example.com")
	if err == nil {
		t.Fatal("browserCommand with an unsupported GOOS should return an error, not a guess")
	}
}
