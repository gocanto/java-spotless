package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"jave-spotless/pkg/spotless"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 0
	}

	switch args[0] {
	case "help", "-h", "--help":
		printUsage()
		return 0
	case "check":
		return runMode(spotless.ModeCheck, args[1:])
	case "apply":
		return runMode(spotless.ModeApply, args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", args[0])
		printUsage()
		return 2
	}
}

func runMode(mode spotless.Mode, args []string) int {
	fs := flag.NewFlagSet(string(mode), flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	baseBranch := fs.String("base-branch", "", "Target base branch used to detect changed Java files")
	all := fs.Bool("all", false, "Run against all Java files in the project")
	repoDir := fs.String("repo-dir", ".", "Repository directory to run Spotless against")
	offline := fs.Bool("offline", true, "Use Maven offline mode for local Docker-based runs")
	runner := fs.String("runner", string(spotless.RunnerAuto), "Runner mode: auto, docker, maven")
	remote := fs.String("remote", "origin", "Git remote name used with --base-branch")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "unexpected arguments: %s\n", strings.Join(fs.Args(), " "))
		return 2
	}

	result, err := spotless.Run(context.Background(), spotless.Options{
		Mode:       mode,
		BaseBranch: *baseBranch,
		All:        *all,
		RepoDir:    *repoDir,
		Offline:    *offline,
		Runner:     spotless.RunnerMode(*runner),
		Remote:     *remote,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if result.Skipped {
		fmt.Println("No Java files changed compared to", *baseBranch, "skipping", mode)
		return 0
	}

	if len(result.ChangedFiles) > 0 {
		fmt.Println("Changed Java files:")
		for _, f := range result.ChangedFiles {
			fmt.Printf("  - %s\n", f)
		}
	}

	return 0
}

func printUsage() {
	fmt.Println("spotless - run Spotless code-style checks/applies")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  spotless check --base-branch <branch> [flags]")
	fmt.Println("  spotless apply --base-branch <branch> [flags]")
	fmt.Println("  spotless check --all [flags]")
	fmt.Println("  spotless apply --all [flags]")
	fmt.Println()
	fmt.Println("Common flags:")
	fmt.Println("  --repo-dir <path>       Repository path (default: .)")
	fmt.Println("  --offline=<bool>        Offline mode for local Docker run (default: true)")
	fmt.Println("  --runner <auto|docker|maven>")
	fmt.Println("  --remote <name>         Git remote for base comparison (default: origin)")
	fmt.Println("  --all                   Run against the entire project")
	fmt.Println()
	fmt.Println("Rules:")
	fmt.Println("  --all and --base-branch are mutually exclusive.")
}
