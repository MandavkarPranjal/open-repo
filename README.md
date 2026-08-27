# open-repo

A small cross-platform Go replacement for the `open-repo` Bash script. It
reads fetch remotes from the current Git repository, lets you choose a remote
and web frontend, then opens the resulting repository URL.

## Build and install

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
go test ./...
```
