package repo_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jallum/beadwork/internal/repo"
	"github.com/jallum/beadwork/internal/testutil"
)

func TestGetGitContextMainWorktree(t *testing.T) {
	env := testutil.NewEnv(t)
	defer env.Cleanup()

	ctx := env.Repo.GetGitContext()
	if ctx.IsWorktree {
		t.Fatalf("IsWorktree = true in main working tree")
	}
	if ctx.Branch == "" {
		t.Fatalf("Branch should be populated")
	}
}

func TestGetGitContextLinkedWorktreeFromSubdirectory(t *testing.T) {
	env := testutil.NewEnv(t)
	defer env.Cleanup()

	wtDir := filepath.Join(t.TempDir(), "linked")
	runGit(t, env.Dir, "worktree", "add", "-b", "feature/context", wtDir)

	subdir := filepath.Join(wtDir, "nested", "dir")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	r, err := repo.FindRepoAt(subdir)
	if err != nil {
		t.Fatalf("FindRepoAt linked worktree subdir: %v", err)
	}

	ctx := r.GetGitContext()
	if !ctx.IsWorktree {
		t.Fatalf("IsWorktree = false from linked worktree subdirectory")
	}
	if ctx.Branch != "feature/context" {
		t.Fatalf("Branch = %q, want feature/context", ctx.Branch)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %s: %v", args, out, err)
	}
}
