package spotless

import (
	"context"
	"strings"
	"testing"
)

type call struct {
	dir  string
	env  map[string]string
	name string
	args []string
}

type fakeExecutor struct {
	runCalls    []call
	outputCalls []call
	outputs     []string
}

func (f *fakeExecutor) Run(_ context.Context, dir string, env map[string]string, name string, args ...string) error {
	f.runCalls = append(f.runCalls, call{dir: dir, env: env, name: name, args: append([]string(nil), args...)})
	return nil
}

func (f *fakeExecutor) Output(_ context.Context, dir string, env map[string]string, name string, args ...string) (string, error) {
	f.outputCalls = append(f.outputCalls, call{dir: dir, env: env, name: name, args: append([]string(nil), args...)})
	if len(f.outputs) == 0 {
		return "", nil
	}
	out := f.outputs[0]
	f.outputs = f.outputs[1:]
	return out, nil
}

func TestApplyAllOmitsSpotlessFilesArg(t *testing.T) {
	ex := &fakeExecutor{}
	res, err := runWithExecutor(context.Background(), Options{
		Mode:    ModeApply,
		All:     true,
		RepoDir: ".",
		Runner:  RunnerMaven,
	}, ex)
	if err != nil {
		t.Fatalf("runWithExecutor returned error: %v", err)
	}
	if res.Skipped {
		t.Fatalf("expected non-skipped result")
	}
	if len(ex.runCalls) != 1 {
		t.Fatalf("expected 1 run call, got %d", len(ex.runCalls))
	}
	for _, arg := range ex.runCalls[0].args {
		if strings.HasPrefix(arg, "-DspotlessFiles=") {
			t.Fatalf("did not expect spotlessFiles arg for --all, got %q", arg)
		}
	}
}

func TestCheckAllOmitsSpotlessFilesArg(t *testing.T) {
	ex := &fakeExecutor{}
	_, err := runWithExecutor(context.Background(), Options{
		Mode:    ModeCheck,
		All:     true,
		RepoDir: ".",
		Runner:  RunnerMaven,
	}, ex)
	if err != nil {
		t.Fatalf("runWithExecutor returned error: %v", err)
	}
	if len(ex.runCalls) != 1 {
		t.Fatalf("expected 1 run call, got %d", len(ex.runCalls))
	}
	for _, arg := range ex.runCalls[0].args {
		if strings.HasPrefix(arg, "-DspotlessFiles=") {
			t.Fatalf("did not expect spotlessFiles arg for --all, got %q", arg)
		}
	}
}

func TestApplyWithBaseBranchUsesChangedFilesPattern(t *testing.T) {
	ex := &fakeExecutor{outputs: []string{"src/main/java/Foo.java\n", "", ""}}
	res, err := runWithExecutor(context.Background(), Options{
		Mode:       ModeApply,
		All:        false,
		BaseBranch: "main",
		RepoDir:    ".",
		Runner:     RunnerMaven,
	}, ex)
	if err != nil {
		t.Fatalf("runWithExecutor returned error: %v", err)
	}
	if len(res.ChangedFiles) != 1 || res.ChangedFiles[0] != "src/main/java/Foo.java" {
		t.Fatalf("unexpected changed files: %#v", res.ChangedFiles)
	}
	if len(ex.outputCalls) != 3 {
		t.Fatalf("expected 3 git diff calls, got %d", len(ex.outputCalls))
	}
	if len(ex.runCalls) != 1 {
		t.Fatalf("expected 1 run call, got %d", len(ex.runCalls))
	}
	found := false
	for _, arg := range ex.runCalls[0].args {
		if strings.HasPrefix(arg, "-DspotlessFiles=") {
			found = true
			if !strings.Contains(arg, "src/main/java/Foo\\.java") {
				t.Fatalf("unexpected spotlessFiles pattern: %q", arg)
			}
		}
	}
	if !found {
		t.Fatalf("expected spotlessFiles arg for changed-files mode")
	}
}

func TestAllAndBaseBranchAreMutuallyExclusive(t *testing.T) {
	ex := &fakeExecutor{}
	_, err := runWithExecutor(context.Background(), Options{
		Mode:       ModeCheck,
		All:        true,
		BaseBranch: "main",
		RepoDir:    ".",
		Runner:     RunnerMaven,
	}, ex)
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("unexpected error: %v", err)
	}
}
