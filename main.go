package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

var version = "dev"

type options struct {
	remote, frontend string
	dryRun, printURL bool
}

func main() {
	opts, runProgram, err := parseOptions(os.Args[1:], os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if !runProgram {
		return
	}
	if err := run(opts, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseOptions(args []string, stdout, stderr io.Writer) (options, bool, error) {
	var opts options
	flags := flag.NewFlagSet("open-repo", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&opts.remote, "remote", "", "remote name to open")
	flags.StringVar(&opts.frontend, "frontend", "", "frontend name to use")
	flags.BoolVar(&opts.dryRun, "dry-run", false, "print what would be opened without launching a browser")
	flags.BoolVar(&opts.printURL, "print-url", false, "print the repository URL without launching a browser")
	showHelp := flags.Bool("help", false, "show this help message and exit")
	showVersion := flags.Bool("version", false, "print the version and exit")
	flags.Usage = func() {
		_, _ = fmt.Fprint(stderr, "Usage: open-repo [--remote NAME] [--frontend NAME] [--dry-run|--print-url]\n\n")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return opts, false, nil
		}
		return opts, false, err
	}
	if flags.NArg() != 0 {
		return opts, false, fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	if *showHelp {
		flags.Usage()
		return opts, false, nil
	}
	if *showVersion {
		if _, err := fmt.Fprintln(stdout, version); err != nil {
			return opts, false, err
		}
		return opts, false, nil
	}
	return opts, true, nil
}

func run(opts options, stdout io.Writer) error {
	if dir := tmuxRepositoryDir(); dir != "" {
		if err := os.Chdir(dir); err != nil {
			return fmt.Errorf("change to tmux repository directory: %w", err)
		}
	}

	remotes, err := gitRemotes()
	if err != nil {
		return err
	}
	if len(remotes) == 0 {
		return errors.New("no git remotes found")
	}

	selected, err := selectedRemote(remotes, opts.remote)
	if err != nil {
		return err
	}
	repoPath, err := repositoryPath(selected.URL)
	if err != nil {
		return err
	}

	destination, err := selectedFrontend(opts.frontend)
	if err != nil {
		return err
	}

	url, err := frontendURL(destination, repoPath)
	if err != nil {
		return err
	}
	if opts.dryRun || opts.printURL {
		_, err := fmt.Fprintln(stdout, url)
		return err
	}
	if _, err := fmt.Fprintf(stdout, "Opening: %s\n", url); err != nil {
		return err
	}
	return launchBrowser(url)
}

func selectedRemote(remotes []remote, name string) (remote, error) {
	if name == "" {
		selected, ok, err := chooseRemote(remotes)
		if err != nil || !ok {
			return remote{}, err
		}
		return selected, nil
	}
	for _, remote := range remotes {
		if remote.Name == name {
			return remote, nil
		}
	}
	return remote{}, fmt.Errorf("remote %q not found", name)
}

func selectedFrontend(name string) (string, error) {
	if name == "" {
		frontend, ok, err := selectItem("Open in: ", frontends, inTmuxPopup())
		if err != nil || !ok {
			return "", err
		}
		return frontend, nil
	}
	if _, err := frontendURL(name, "owner/repo"); err != nil {
		return "", err
	}
	return name, nil
}
