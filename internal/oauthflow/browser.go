package oauthflow

import (
	"fmt"
	"os/exec"
	"runtime"
)

// OpenBrowser best-effort launches the system browser at url. Failure here
// isn't fatal to the login flow — the caller should still display the URL
// so the user can open it manually.
func OpenBrowser(url string) error {
	name, args, err := browserCommand(runtime.GOOS, url)
	if err != nil {
		return err
	}
	return exec.Command(name, args...).Start()
}

// browserCommand returns the binary + args to launch url in the system
// browser for the given GOOS. Separated from OpenBrowser so the
// OS-dependent command construction is testable without actually spawning a
// process.
func browserCommand(goos, url string) (string, []string, error) {
	switch goos {
	case "darwin":
		return "open", []string{url}, nil
	case "linux":
		return "xdg-open", []string{url}, nil
	case "windows":
		return "rundll32", []string{"url.dll,FileProtocolHandler", url}, nil
	default:
		return "", nil, fmt.Errorf("don't know how to open a browser on GOOS=%q", goos)
	}
}
