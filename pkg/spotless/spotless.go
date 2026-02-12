package spotless

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Mode string

const (
	ModeCheck Mode = "check"
	ModeApply Mode = "apply"
)

type RunnerMode string

const (
	RunnerAuto   RunnerMode = "auto"
	RunnerDocker RunnerMode = "docker"
	RunnerMaven  RunnerMode = "maven"
)

type Options struct {
	Mode       Mode
	BaseBranch string
	All        bool
	RepoDir    string
	Offline    bool
	Runner     RunnerMode
	Remote     string
}

type Result struct {
	ChangedFiles []string
	CommandLine  []string
	UsedRunner   RunnerMode
	Skipped      bool
}

type executor interface {
	Run(ctx context.Context, dir string, env map[string]string, name string, args ...string) error
	Output(ctx context.Context, dir string, env map[string]string, name string, args ...string) (string, error)
}

type systemExecutor struct{}

func (s systemExecutor) Run(ctx context.Context, dir string, env map[string]string, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = mergedEnv(env)
	return cmd.Run()
}

func (s systemExecutor) Output(ctx context.Context, dir string, env map[string]string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = mergedEnv(env)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func Run(ctx context.Context, opts Options) (Result, error) {
	return runWithExecutor(ctx, opts, systemExecutor{})
}

func runWithExecutor(ctx context.Context, opts Options, ex executor) (Result, error) {
	opts = normalize(opts)
	if err := validate(opts); err != nil {
		return Result{}, err
	}

	usedRunner := pickRunner(opts)
	result := Result{UsedRunner: usedRunner}

	var spotlessFilesArg string
	if !opts.All {
		changed, err := changedJavaFiles(ctx, ex, opts.RepoDir, opts.Remote, opts.BaseBranch)
		if err != nil {
			return Result{}, fmt.Errorf("collect changed java files: %w", err)
		}
		if len(changed) == 0 {
			result.Skipped = true
			return result, nil
		}
		result.ChangedFiles = changed
		spotlessFilesArg = "-DspotlessFiles=" + buildSpotlessFilesPattern(changed)
	}

	if usedRunner == RunnerMaven {
		args := []string{"--batch-mode", "--errors", "--fail-at-end", "-Dmaven.repo.local=/m2/repository", "spotless:" + string(opts.Mode)}
		if spotlessFilesArg != "" {
			args = append(args, spotlessFilesArg)
		}
		result.CommandLine = append([]string{"mvn"}, args...)
		if err := ex.Run(ctx, opts.RepoDir, nil, "mvn", args...); err != nil {
			return result, err
		}
		return result, nil
	}

	composeCmd, err := detectCompose(ctx, ex)
	if err != nil {
		return result, err
	}
	repoDirAbs, err := filepath.Abs(opts.RepoDir)
	if err != nil {
		repoDirAbs = opts.RepoDir
	}
	service := "spotless-" + string(opts.Mode)
	args := make([]string, 0, 16)
	if len(composeCmd) == 2 {
		args = append(args, composeCmd[1])
	}
	args = append(args, "-f", "docker-compose.spotless.yml", "run", "--rm", service, "mvn")
	if opts.Offline {
		args = append(args, "-o")
	}
	args = append(args, "-Dmaven.repo.local=/m2/repository", "spotless:"+string(opts.Mode))
	if spotlessFilesArg != "" {
		args = append(args, spotlessFilesArg)
	}

	env := map[string]string{
		"HOST_UID":               strconv.Itoa(os.Getuid()),
		"HOST_GID":               strconv.Itoa(os.Getgid()),
		"CODE_STYLE_ENV_ENABLED": "true",
		"REPO_DIR":               repoDirAbs,
	}
	result.CommandLine = append([]string{composeCmd[0]}, args...)
	if err := ex.Run(ctx, ".", env, composeCmd[0], args...); err != nil {
		return result, err
	}
	return result, nil
}

func normalize(opts Options) Options {
	if opts.RepoDir == "" {
		opts.RepoDir = "."
	}
	if opts.Remote == "" {
		opts.Remote = "origin"
	}
	if opts.Runner == "" {
		opts.Runner = RunnerAuto
	}
	return opts
}

func validate(opts Options) error {
	if opts.Mode != ModeCheck && opts.Mode != ModeApply {
		return fmt.Errorf("invalid mode %q", opts.Mode)
	}
	if opts.Runner != RunnerAuto && opts.Runner != RunnerDocker && opts.Runner != RunnerMaven {
		return fmt.Errorf("invalid runner %q (valid: auto, docker, maven)", opts.Runner)
	}
	if opts.All && opts.BaseBranch != "" {
		return errors.New("--all and --base-branch are mutually exclusive")
	}
	if !opts.All && opts.BaseBranch == "" {
		return errors.New("--base-branch is required unless --all is used")
	}
	return nil
}

func pickRunner(opts Options) RunnerMode {
	if opts.Runner != RunnerAuto {
		return opts.Runner
	}
	if os.Getenv("CI") != "" {
		return RunnerMaven
	}
	return RunnerDocker
}

func changedJavaFiles(ctx context.Context, ex executor, repoDir, remote, base string) ([]string, error) {
	queries := [][]string{
		{"diff", "--name-only", "--diff-filter=d", fmt.Sprintf("%s/%s...HEAD", remote, base), "--", "*.java"},
		{"diff", "--cached", "--name-only", "--diff-filter=d", "--", "*.java"},
		{"diff", "--name-only", "--diff-filter=d", "--", "*.java"},
	}

	set := map[string]struct{}{}
	for _, q := range queries {
		out, err := ex.Output(ctx, repoDir, nil, "git", q...)
		if err != nil {
			// Keep compatibility with the previous Makefile behavior where
			// diff failures (e.g. missing origin/<base>) were treated as empty.
			continue
		}
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || !strings.HasSuffix(line, ".java") {
				continue
			}
			set[line] = struct{}{}
		}
	}

	files := make([]string, 0, len(set))
	for f := range set {
		files = append(files, f)
	}
	sort.Strings(files)
	return files, nil
}

func buildSpotlessFilesPattern(files []string) string {
	parts := make([]string, 0, len(files))
	for _, file := range files {
		parts = append(parts, regexp.QuoteMeta(filepath.ToSlash(file)))
	}
	return "(" + strings.Join(parts, "|") + ")"
}

func detectCompose(ctx context.Context, ex executor) ([]string, error) {
	if _, err := ex.Output(ctx, ".", nil, "docker-compose", "version"); err == nil {
		return []string{"docker-compose"}, nil
	}
	if _, err := ex.Output(ctx, ".", nil, "docker", "compose", "version"); err == nil {
		return []string{"docker", "compose"}, nil
	}
	return nil, errors.New("docker compose command not found; install docker-compose or docker compose")
}

func mergedEnv(extra map[string]string) []string {
	base := os.Environ()
	if len(extra) == 0 {
		return base
	}
	kv := make(map[string]string, len(base)+len(extra))
	for _, entry := range base {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) == 2 {
			kv[parts[0]] = parts[1]
		}
	}
	for k, v := range extra {
		kv[k] = v
	}
	out := make([]string, 0, len(kv))
	for k, v := range kv {
		out = append(out, k+"="+v)
	}
	sort.Strings(out)
	return out
}
