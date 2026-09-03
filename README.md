# faro

`faro` — the terminal client for **Faro Helm** (standups, attendance, leave). No more browser tab for daily standups.

This is an early release — rough edges expected. Bug reports welcome.

## Install

### macOS

```bash
brew tap beamlabco/tap
brew install faro
```

> **Heads up:** if `brew install` fails with `Your Command Line Tools are too outdated`, update them first — System Settings → General → Software Update, or:
> ```bash
> sudo rm -rf /Library/Developer/CommandLineTools
> sudo xcode-select --install
> ```
> Then retry the install. This is a one-time machine thing, not a `faro` issue.

### Linux

If you already have Homebrew (a.k.a. Linuxbrew) set up, the same two commands work as-is:
```bash
brew tap beamlabco/tap
brew install faro
```

No Homebrew? Grab the prebuilt binary from the [releases page](https://github.com/beamlabco/faro-helm-cli/releases) instead — `faro_<version>_linux_amd64.tar.gz` (or `linux_arm64` on an ARM machine):
```bash
curl -LO https://github.com/beamlabco/faro-helm-cli/releases/latest/download/faro_0.0.1_linux_amd64.tar.gz
tar -xzf faro_0.0.1_linux_amd64.tar.gz
sudo mv faro /usr/local/bin/faro
faro --version
```

### Windows

No Homebrew equivalent yet — grab `faro_<version>_windows_amd64.zip` from the [releases page](https://github.com/beamlabco/faro-helm-cli/releases), unzip it, and put `faro.exe` somewhere on your `PATH` (or run it from wherever you unzipped it). In PowerShell:
```powershell
Invoke-WebRequest -Uri https://github.com/beamlabco/faro-helm-cli/releases/latest/download/faro_0.0.1_windows_amd64.zip -OutFile faro.zip
Expand-Archive faro.zip -DestinationPath .
.\faro.exe --version
```
Windows arm64 isn't built yet — see "What's not in this release yet" below.

Confirm it worked (macOS/Linux):
```bash
faro --version
```

## First run

```bash
faro
```
drops you into an interactive shell talking to the production Helm API. Type `login` — it'll give you a device code + URL to approve in your browser, then you're in.

**First time / new to a workspace?** If you've received an invitation, type `join` instead — it'll prompt for your invitation token (paste it in), name, email, and a password to set, and creates + logs you in in one step. Existing members just use `login`.

## Commands

```
standup            submit today's standup
standup today       view today's standup
standup history     past standups

checkin / checkout  attendance
attendance today
attendance history

leave               request leave (shows your live quota/balance first)
leave list
leave cancel

team                your teams
people              workspace people
password            change password
whoami
help
clear
quit
```

Every command here is available to every logged-in user — there's no separate admin/manager tier in the CLI. Manager stuff (approvals, team setup, etc.) still lives in the web app for now.

## Config

Lives at `~/.faro-helm/config.yaml`. Override the API endpoint with:
```bash
export FARO_HELM_API_URL=https://api.farohelm.com   # default in the release build
```

## Building from source

```bash
make build         # dev build → localhost:3001
make build-staging # staging build
make build-prod    # production build
make build-all     # cross-platform production builds (mac/linux/windows)
make install       # copy to /usr/local/bin (production build)
go test ./...
```

## Found a bug / want a feature?

Open an issue: **https://github.com/beamlabco/faro-helm-cli/issues**
(bug / feature / chore templates are set up)

## What's not in this release yet

- Windows arm64 build (windows amd64, and both mac/linux archs, are covered)
- No `--help` flag parsing beyond `--version` — everything else is inside the interactive shell

Feedback welcome 🙏
