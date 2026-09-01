# open-repo

A small cross-platform Go replacement for the [`open-repo`](https://raw.githubusercontent.com/MandavkarPranjal/environment/refs/heads/main/home/scripts/open-repo) Bash script. It
reads fetch remotes from the current Git repository, lets you choose a remote
and web frontend (including GitHub and GitLab), then opens the resulting
repository URL.

## Install

If Go is installed, install the latest version directly from GitHub:

```sh
go install github.com/MandavkarPranjal/open-repo@latest
```

This installs `open-repo` into Go's binary directory (`$GOBIN`, or usually
`$HOME/go/bin`). Make sure that directory is on your `PATH`.

Prebuilt binaries for Linux, macOS, and Windows are available on the
[Releases page](https://github.com/MandavkarPranjal/open-repo/releases).

## Build locally

This project uses [mise](https://mise.jdx.dev/) to pin the Go toolchain used
for development. Install mise, then run this from the repository root:

```sh
mise install
```

The pinned Go version is `1.27.0`. Available development commands are:

```sh
mise run build       # Build the binary
mise run test        # Run all tests
mise run check       # Check formatting, tests, and static analysis
mise run fmt         # Format Go source files
mise run fmt-check   # Fail if formatting is needed
mise run vet         # Run static analysis
mise run lint        # Run golangci-lint
mise run staticcheck # Run Staticcheck
mise run run         # Run open-repo from source
mise run install     # Install open-repo with go install
mise run tidy        # Update go.mod and go.sum
```

## Usage

Run `open-repo` interactively, or specify a remote and frontend for scripts:

```sh
open-repo --remote origin --frontend GitHub --print-url
open-repo --remote origin --frontend GitLab --dry-run
open-repo --version
```

`--print-url` and `--dry-run` print the generated URL and do not start a
browser. Run `open-repo --help` for the complete flag list.

### Browser configuration

`open-repo` first respects `$BROWSER`; it may include command-line arguments,
for example `BROWSER="firefox --private-window"`. To configure a persistent
browser command, create `~/.config/open-repo/config.toml`:

```toml
browser = "firefox --private-window"
```

`$BROWSER` takes precedence over this configuration. If neither is set,
open-repo uses the platform's standard opener.

You can also run the default Go commands directly after `mise install`:

```sh
go build -o open-repo .
install -Dm755 open-repo ~/.local/bin/open-repo
```

Run `open-repo` from a Git repository. If `fzf` is installed it is used for
both selectors; otherwise the CLI presents numbered choices. Browser opening
uses `omarchy-launch-browser`, `gio`, or `xdg-open` on Linux, `open` on macOS,
and Windows' `rundll32` URL handler. The Zen/uwsm path is preserved on Linux.

Cross-compile examples:

```sh
GOOS=darwin GOARCH=arm64 go build -o open-repo .
GOOS=windows GOARCH=amd64 go build -o open-repo.exe .
```

## Verify

```sh
mise run check
```

To verify a downloaded release archive, download its matching `checksums.txt`
asset from the same GitHub release and check the archive listed by GoReleaser:

```sh
tag=vX.Y.Z
archive=open-repo_linux_amd64.tar.gz
base="https://github.com/MandavkarPranjal/open-repo/releases/download/$tag"
curl -fLO "$base/$archive"
curl -fLO "$base/checksums.txt"
grep -F "  $archive" checksums.txt | sha256sum -c -
```
