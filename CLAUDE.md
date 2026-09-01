# faro-helm-cli

Go CLI for **Faro Helm** — the `faro` binary. Built with Bubble Tea (TUI).
Talks to `faro-helm-api` and `faro-auth-api` over REST. All endpoints use the `/api/v1/` prefix.

## Binary

```bash
faro standup
faro checkin
faro leave list
# etc.
```

Config stored at `~/.faro-helm/config.yaml`.

| Variable | Default | Purpose |
|---|---|---|
| `FARO_HELM_API_URL` | `http://localhost:3001` | Helm API base URL (overrides build-time ldflag) |
| `FARO_AUTH_API_URL` | `http://localhost:3000` | Auth API base URL (overrides build-time ldflag) |

Helm calls go to `{baseURL}/api/v1/*`. Auth is split: `/login` uses OAuth 2.0 Authorization Code + PKCE against `{authBaseURL}/oauth/*` (root-level, unversioned); register/join/change-password/me still use the legacy `{authBaseURL}/api/v1/auth/*` endpoints, unchanged until those get their own migration.

## Login (`/login`)

`faro-helm-cli` is a registered public OAuth 2.0 client (`faro-helm-cli`, no client secret — PKCE instead). `/login` opens the system browser to `{authBaseURL}/oauth/authorize`, where the user actually enters their email/password (the CLI never sees the password) — a local loopback server on an ephemeral port catches the redirect, and the terminal shows the URL as a manual fallback in case auto-open fails.

There is no device-code/browserless flow — the CLI is assumed to always run somewhere a browser is reachable, so Authorization Code + PKCE covers every case with one flow. Package `internal/oauthflow` is the framework-free core: PKCE generation, the loopback callback server, authorize-URL building, and cross-platform browser launching; `internal/auth.Service.BeginBrowserLogin`/`CompleteBrowserLogin` wire it to token exchange and config persistence; `internal/ui.BrowserLoginModel` is the Bubble Tea screen.

`internal/oauthflow` has unit tests (including a real local HTTP server exercised over the network, run with `-race`); `internal/api`'s token-exchange call is tested against an `httptest.Server`. The Bubble Tea layer itself isn't unit tested, matching this repo's existing convention for `ui/*.go`.

## Commands

```bash
make build         # build ./bin/faro (dev → localhost:3001)
make build-staging # build (→ helm-faro.beamlab.dev)
make build-prod    # build (→ api.helm.farohelm.com)
make run           # go run (dev)
make run-staging   # go run (staging)
make run-prod      # go run (production)
make install       # copy to /usr/local/bin/faro (production build)
make build-all     # cross-platform production binaries
go test ./...
go fmt ./...
```

## Source layout

```
cmd/faro/main.go        entry point — wires services, starts Bubble Tea program
internal/
├── config/             loads ~/.faro-helm/config.yaml, FARO_HELM_API_URL env
├── api/                HTTP client (resty) for all API calls
├── oauthflow/          PKCE, loopback callback server, authorize-URL building, browser launch — no Bubble Tea or config dependency
├── auth/               login (browser + PKCE), logout, register, join
├── standup/            submit, list, history
├── attendance/         checkin, checkout, today, history
├── leave/              request (leave types + quota lookup), list, cancel
├── project/            list (own projects)
├── user/               team list, password
└── ui/                 Bubble Tea screens + styles
    ├── shell.go        main REPL shell + dashboard
    ├── styles.go       teal color palette (#1D9E75 primary, #5DCAA5 secondary)
    ├── commands.go     command registry + autocomplete
    └── *.go            one file per screen
```

This is a **member/self-service CLI only** — the same scope as `faro-helm-app`. It exposes no primary/manager administration (org settings, invitations, role changes, project CRUD, project membership, office-hours overrides, password resets, or leave review). Those remain server-side in `faro-helm-api` for other clients; this binary simply doesn't expose them.

## Module

`github.com/beamlabco/faro-helm`

## CLI Commands

All commands are available to every authenticated user — there is no role-gated
admin/manager tier in this CLI.

`standup`, `standup today`, `standup history`, `checkin`, `checkout`,
`attendance today`, `attendance history`, `leave`, `leave list`,
`leave cancel`, `project`, `team`, `password`, `whoami`, `help`, `clear`, `quit`

There is no standalone `attendance` (mark) command — `checkin`/`checkout` are the only
way to record attendance; `faro-helm-api` dropped the generic mark-attendance endpoint.

`leave` (request) fetches the workspace's active leave types and the caller's quota/used/
remaining for each via `GET /leaves/balance` before prompting for dates — leave types are
workspace-configurable, not a fixed enum, so the type picker is always live-loaded.

## TUI colors (teal palette)

- Primary: `#1D9E75`
- Secondary/muted accent: `#5DCAA5`
- Error: `#EF4444`
- Success: `#10B981`
- Muted: `#6B7280`
