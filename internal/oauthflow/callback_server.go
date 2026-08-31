package oauthflow

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

// CallbackServer is the loopback HTTP server that catches the browser
// redirect at the end of the Authorization Code flow. It binds an ephemeral
// local port — matching the 'http://localhost:*/callback' wildcard
// registered for faro-helm-cli in core.oauth_clients.
type CallbackServer struct {
	listener net.Listener
	server   *http.Server
	resultCh chan callbackResult
}

type callbackResult struct {
	code  string
	state string
	err   error
}

// StartCallbackServer binds 127.0.0.1 on an OS-assigned free port and starts
// serving immediately. Call Close when done with it.
func StartCallbackServer() (*CallbackServer, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to bind a local port for the OAuth callback: %w", err)
	}

	s := &CallbackServer{
		listener: ln,
		resultCh: make(chan callbackResult, 1),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", s.handleCallback)
	s.server = &http.Server{Handler: mux}

	go s.server.Serve(ln) //nolint:errcheck // Close() below is expected to produce http.ErrServerClosed

	return s, nil
}

// RedirectURI is the exact redirect_uri to send in the authorize request and
// the token exchange — they must match.
func (s *CallbackServer) RedirectURI() string {
	return fmt.Sprintf("http://%s/callback", s.listener.Addr().String())
}

// Await blocks until the browser hits the callback URL (or ctx is done) and
// returns the authorization code. A response with a mismatched state or an
// error/error_description param is surfaced as an error, not a bare code.
func (s *CallbackServer) Await(ctx context.Context, expectedState string) (string, error) {
	select {
	case r := <-s.resultCh:
		if r.err != nil {
			return "", r.err
		}
		if r.state != expectedState {
			return "", fmt.Errorf("state mismatch on OAuth callback (possible CSRF): expected %q, got %q", expectedState, r.state)
		}
		return r.code, nil
	case <-ctx.Done():
		return "", fmt.Errorf("timed out waiting for browser sign-in: %w", ctx.Err())
	}
}

// Close shuts down the local server and releases the port.
func (s *CallbackServer) Close() error {
	return s.server.Close()
}

func (s *CallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	if errParam := q.Get("error"); errParam != "" {
		msg := errParam
		if desc := q.Get("error_description"); desc != "" {
			msg = errParam + ": " + desc
		}
		s.deliver(callbackResult{err: fmt.Errorf("%s", msg)})
		writeCallbackPage(w, false)
		return
	}

	s.deliver(callbackResult{code: q.Get("code"), state: q.Get("state")})
	writeCallbackPage(w, true)
}

// deliver is a non-blocking send — only the first callback hit is kept.
// Later hits (a browser retry, a favicon request, a double-click) must not
// block the HTTP handler or panic on a second send to a full channel.
func (s *CallbackServer) deliver(r callbackResult) {
	select {
	case s.resultCh <- r:
	default:
	}
}

func writeCallbackPage(w http.ResponseWriter, ok bool) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if ok {
		fmt.Fprint(w, `<!DOCTYPE html><html><body style="font-family:sans-serif;text-align:center;margin-top:20vh">
<h2>Signed in to Faro Helm</h2><p>You can close this tab and return to your terminal.</p></body></html>`)
		return
	}
	fmt.Fprint(w, `<!DOCTYPE html><html><body style="font-family:sans-serif;text-align:center;margin-top:20vh">
<h2>Sign-in failed</h2><p>Return to your terminal for details. You can close this tab.</p></body></html>`)
}
