# faro-helm-cli

Go CLI for **Faro Helm** — the `faro` binary. Built with Bubble Tea (TUI).
Talks to `faro-helm-api` over REST.

## Binary

```bash
faro standup
faro checkin
faro leave list
# etc.
```

Config stored at `~/.faro-helm/config.yaml`.
API URL set via build-time ldflags or `FARO_HELM_API_URL` env var.

## Commands

```bash
make build        # build ./bin/faro (dev → localhost:3001)
make build-prod   # build (→ api.helm.farohelm.com)
make run          # go run (dev)
make install      # copy to /usr/local/bin/faro
make build-all    # cross-platform binaries
go test ./...
go fmt ./...
```

## Source layout

```
cmd/faro/main.go        entry point — wires services, starts Bubble Tea program
internal/
├── config/             loads ~/.faro-helm/config.yaml, FARO_HELM_API_URL env
├── api/                HTTP client (resty) for all API calls
├── auth/               login, logout, register, join
├── standup/            submit, list, history
├── attendance/         mark, checkin, checkout, history
├── leave/              request, list, review, cancel
├── project/            list, create, settings, members
├── invitation/         create, accept
├── organization/       get/update settings
├── user/               list, role, office hours, password
└── ui/                 Bubble Tea screens + styles
    ├── shell.go        main REPL shell + dashboard
    ├── styles.go       teal color palette (#1D9E75 primary, #5DCAA5 secondary)
    ├── commands.go     command registry + autocomplete
    └── *.go            one file per screen
```

## Module

`github.com/beamlabco/faro-helm`

## CLI Commands

### All users
`standup`, `standup today`, `standup history`, `checkin`, `checkout`,
`attendance`, `attendance today`, `attendance history`, `leave`, `leave list`,
`leave cancel`, `project`, `team`, `password`, `whoami`, `help`, `clear`, `quit`

### Primary only
`project create`, `project settings`, `project members`, `role`, `office-hours`,
`invite`, `settings`, `reset-password`

### Primary + Manager
`leave review`

## TUI colors (teal palette)

- Primary: `#1D9E75`
- Secondary/muted accent: `#5DCAA5`
- Error: `#EF4444`
- Success: `#10B981`
- Muted: `#6B7280`
