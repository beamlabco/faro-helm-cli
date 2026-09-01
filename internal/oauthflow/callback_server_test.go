package oauthflow

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCallbackServer_RedirectURIFormat(t *testing.T) {
	s, err := StartCallbackServer()
	if err != nil {
		t.Fatalf("StartCallbackServer() error: %v", err)
	}
	defer s.Close()

	uri := s.RedirectURI()
	if !strings.HasPrefix(uri, "http://127.0.0.1:") || !strings.HasSuffix(uri, "/callback") {
		t.Errorf("RedirectURI() = %q, want http://127.0.0.1:<port>/callback", uri)
	}
}

func TestCallbackServer_DeliversCodeOnMatchingState(t *testing.T) {
	s, err := StartCallbackServer()
	if err != nil {
		t.Fatalf("StartCallbackServer() error: %v", err)
	}
	defer s.Close()

	// Collect the simulated browser's HTTP outcome on a channel instead of
	// calling t.Errorf from the goroutine — Await() can return as soon as
	// the handler reads the query params, before the HTTP response finishes
	// writing back to this goroutine's http.Get, so the test function could
	// otherwise return (and be marked done) while this goroutine is still
	// running, and testing.T panics on t.Errorf from a finished test.
	type getOutcome struct {
		statusCode int
		err        error
	}
	outcomeCh := make(chan getOutcome, 1)
	go func() {
		resp, err := http.Get(s.RedirectURI() + "?code=the-code&state=xyz")
		if err != nil {
			outcomeCh <- getOutcome{err: err}
			return
		}
		defer resp.Body.Close()
		outcomeCh <- getOutcome{statusCode: resp.StatusCode}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	code, err := s.Await(ctx, "xyz")
	if err != nil {
		t.Fatalf("Await() returned error: %v", err)
	}
	if code != "the-code" {
		t.Errorf("Await() code = %q, want %q", code, "the-code")
	}

	outcome := <-outcomeCh
	if outcome.err != nil {
		t.Fatalf("simulated browser redirect failed: %v", outcome.err)
	}
	if outcome.statusCode != http.StatusOK {
		t.Errorf("callback page returned %d, want 200", outcome.statusCode)
	}
}

func TestCallbackServer_RejectsMismatchedState(t *testing.T) {
	s, err := StartCallbackServer()
	if err != nil {
		t.Fatalf("StartCallbackServer() error: %v", err)
	}
	defer s.Close()

	go func() {
		time.Sleep(20 * time.Millisecond)
		http.Get(s.RedirectURI() + "?code=the-code&state=WRONG")
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = s.Await(ctx, "expected-state")
	if err == nil {
		t.Fatal("Await() succeeded with a mismatched state, want an error")
	}
}

func TestCallbackServer_SurfacesProviderError(t *testing.T) {
	s, err := StartCallbackServer()
	if err != nil {
		t.Fatalf("StartCallbackServer() error: %v", err)
	}
	defer s.Close()

	go func() {
		time.Sleep(20 * time.Millisecond)
		http.Get(s.RedirectURI() + "?error=access_denied&error_description=user+said+no&state=xyz")
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = s.Await(ctx, "xyz")
	if err == nil {
		t.Fatal("Await() succeeded despite an error param, want an error")
	}
	if !strings.Contains(err.Error(), "access_denied") {
		t.Errorf("error = %q, want it to mention access_denied", err.Error())
	}
}

func TestCallbackServer_AwaitTimesOutWithNoRequest(t *testing.T) {
	s, err := StartCallbackServer()
	if err != nil {
		t.Fatalf("StartCallbackServer() error: %v", err)
	}
	defer s.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = s.Await(ctx, "xyz")
	if err == nil {
		t.Fatal("Await() succeeded with no request ever arriving, want a timeout error")
	}
}

func TestCallbackServer_IgnoresRequestsAfterFirstDelivery(t *testing.T) {
	// A browser retry, favicon fetch, or double-click shouldn't panic or
	// deadlock the server after the first callback has already been
	// delivered and consumed.
	s, err := StartCallbackServer()
	if err != nil {
		t.Fatalf("StartCallbackServer() error: %v", err)
	}
	defer s.Close()

	go func() {
		time.Sleep(10 * time.Millisecond)
		http.Get(s.RedirectURI() + "?code=first&state=xyz")
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := s.Await(ctx, "xyz"); err != nil {
		t.Fatalf("first Await() failed: %v", err)
	}

	// Second hit after the result was already consumed — must not hang the
	// HTTP handler or crash the server.
	resp, err := http.Get(s.RedirectURI() + "?code=second&state=xyz")
	if err != nil {
		t.Fatalf("second request to the callback server failed: %v", err)
	}
	resp.Body.Close()
}
