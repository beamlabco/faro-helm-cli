package config

import "testing"

func withTestConfigDir(t *testing.T) {
	t.Helper()
	SetConfigDirForTesting(t.TempDir())
	t.Cleanup(func() { SetConfigDirForTesting("") })
}

func TestLoad_NoFileReturnsDefaults(t *testing.T) {
	withTestConfigDir(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.IsAuthenticated() {
		t.Error("expected a fresh config to not be authenticated")
	}
	if cfg.API == nil || cfg.API.BaseURL == "" {
		t.Errorf("expected a default base URL, got %+v", cfg.API)
	}
}

func TestSaveAndLoad_RoundTripsAuthData(t *testing.T) {
	withTestConfigDir(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	user := &User{ID: "m1", AccountID: "acc1", Email: "a@b.com", Name: "A", Role: "member"}
	org := &Organization{ID: "ws1", Name: "Acme", Status: "active"}
	cfg.SetAuthData("access-1", "refresh-1", user, org)

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	reloaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if !reloaded.IsAuthenticated() {
		t.Fatal("expected reloaded config to be authenticated")
	}
	if reloaded.Auth.Token != "access-1" || reloaded.Auth.RefreshToken != "refresh-1" {
		t.Errorf("unexpected auth data: %+v", reloaded.Auth)
	}
	if reloaded.User == nil || reloaded.User.Email != "a@b.com" {
		t.Errorf("unexpected user data: %+v", reloaded.User)
	}
	if reloaded.User.AccountID != "acc1" {
		t.Errorf("expected snake_case-tagged field account_id to round-trip, got %q", reloaded.User.AccountID)
	}
}

// TestLoad_DoesNotLeakStateAcrossConfigDirs guards against viper's global
// singleton leaking a Set() from one config dir into a Load() of another —
// exactly the shape of a live process refreshing a token (Save) and then a
// later, unrelated Load() call.
func TestLoad_DoesNotLeakStateAcrossConfigDirs(t *testing.T) {
	SetConfigDirForTesting(t.TempDir())
	first, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	first.SetAuthData("access-1", "refresh-1", &User{Email: "first@b.com"}, &Organization{Name: "First"})
	if err := Save(first); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	SetConfigDirForTesting(t.TempDir())
	t.Cleanup(func() { SetConfigDirForTesting("") })

	second, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if second.IsAuthenticated() {
		t.Errorf("expected a fresh config dir to be unauthenticated, got %+v", second.Auth)
	}
}

func TestClear_RemovesAuthUserAndOrganization(t *testing.T) {
	withTestConfigDir(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	cfg.SetAuthData("access-1", "refresh-1", &User{Email: "a@b.com"}, &Organization{Name: "Acme"})

	cfg.Clear()

	if cfg.IsAuthenticated() {
		t.Error("expected IsAuthenticated() to be false after Clear()")
	}
	if cfg.Auth != nil || cfg.User != nil || cfg.Organization != nil {
		t.Errorf("expected Auth/User/Organization to be nil after Clear(), got Auth=%+v User=%+v Organization=%+v", cfg.Auth, cfg.User, cfg.Organization)
	}
}
