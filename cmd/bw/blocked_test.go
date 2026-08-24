package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/bkono/beadwork/internal/issue"
	"github.com/bkono/beadwork/internal/testutil"
)

func TestCmdBlockedBasic(t *testing.T) {
	env := testutil.NewEnv(t)
	defer env.Cleanup()

	a, _ := env.Store.Create("Blocker", issue.CreateOpts{Priority: intPtr(1)})
	b, _ := env.Store.Create("Blocked task", issue.CreateOpts{Priority: intPtr(2)})
	env.Store.Link(a.ID, b.ID)
	env.Repo.Commit("link")

	var buf bytes.Buffer
	_, err := cmdBlocked(env.Store, []string{}, PlainWriter(&buf), nil)
	if err != nil {
		t.Fatalf("cmdBlocked: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, b.ID) {
		t.Errorf("output should contain blocked issue %s: %q", b.ID, out)
	}
	if !strings.Contains(out, "Blocked task") {
		t.Errorf("output should contain title: %q", out)
	}
	if !strings.Contains(out, a.ID) {
		t.Errorf("output should list blocker %s: %q", a.ID, out)
	}
	if !strings.Contains(out, "Blocker") {
		t.Errorf("output should contain blocker title: %q", out)
	}
	if !strings.Contains(out, "open") {
		t.Errorf("output should include status: %q", out)
	}
}

func TestCmdBlockedResolves(t *testing.T) {
	env := testutil.NewEnv(t)
	defer env.Cleanup()

	a, _ := env.Store.Create("Blocker", issue.CreateOpts{})
	b, _ := env.Store.Create("Blocked", issue.CreateOpts{})
	env.Store.Link(a.ID, b.ID)
	env.Store.Close(a.ID, "")
	env.Repo.Commit("setup")

	var buf bytes.Buffer
	_, err := cmdBlocked(env.Store, []string{}, PlainWriter(&buf), nil)
	if err != nil {
		t.Fatalf("cmdBlocked: %v", err)
	}
	if strings.Contains(buf.String(), b.ID) {
		t.Error("resolved issue should not appear in blocked output")
	}
}

func TestCmdBlockedJSON(t *testing.T) {
	env := testutil.NewEnv(t)
	defer env.Cleanup()

	a, _ := env.Store.Create("Blocker", issue.CreateOpts{})
	b, _ := env.Store.Create("Blocked", issue.CreateOpts{})
	env.Store.Link(a.ID, b.ID)
	env.Repo.Commit("link")

	var buf bytes.Buffer
	_, err := cmdBlocked(env.Store, []string{"--json"}, PlainWriter(&buf), nil)
	if err != nil {
		t.Fatalf("cmdBlocked --json: %v", err)
	}

	var result []struct {
		ID           string `json:"id"`
		Title        string `json:"title"`
		Status       string `json:"status"`
		Priority     int    `json:"priority"`
		OpenBlockers []struct {
			ID     string `json:"id"`
			Title  string `json:"title"`
			Status string `json:"status"`
		} `json:"open_blockers"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("got %d results, want 1", len(result))
	}
	if result[0].ID != b.ID {
		t.Errorf("id = %q, want %q", result[0].ID, b.ID)
	}
	if result[0].Title != "Blocked" {
		t.Errorf("title = %q, want Blocked", result[0].Title)
	}
	if len(result[0].OpenBlockers) != 1 || result[0].OpenBlockers[0].ID != a.ID {
		t.Errorf("open_blockers = %v, want [%s]", result[0].OpenBlockers, a.ID)
	}
	if result[0].OpenBlockers[0].Title != "Blocker" {
		t.Errorf("blocker title = %q, want Blocker", result[0].OpenBlockers[0].Title)
	}
	if result[0].OpenBlockers[0].Status != "open" {
		t.Errorf("blocker status = %q, want open", result[0].OpenBlockers[0].Status)
	}
}

func TestCmdBlockedEmpty(t *testing.T) {
	env := testutil.NewEnv(t)
	defer env.Cleanup()

	env.Store.Create("No deps", issue.CreateOpts{})
	env.Repo.Commit("create")

	var buf bytes.Buffer
	_, err := cmdBlocked(env.Store, []string{}, PlainWriter(&buf), nil)
	if err != nil {
		t.Fatalf("cmdBlocked: %v", err)
	}
	if !strings.Contains(buf.String(), "no blocked issues") {
		t.Errorf("expected 'no blocked issues', got: %q", buf.String())
	}
}

func TestCmdBlockedMultipleBlockers(t *testing.T) {
	env := testutil.NewEnv(t)
	defer env.Cleanup()

	a, _ := env.Store.Create("Blocker A", issue.CreateOpts{Priority: intPtr(1)})
	b, _ := env.Store.Create("Blocker B", issue.CreateOpts{Priority: intPtr(1)})
	c, _ := env.Store.Create("Blocked by two", issue.CreateOpts{Priority: intPtr(2)})
	env.Store.Link(a.ID, c.ID)
	env.Store.Link(b.ID, c.ID)
	env.Repo.Commit("link")

	var buf bytes.Buffer
	_, err := cmdBlocked(env.Store, []string{}, PlainWriter(&buf), nil)
	if err != nil {
		t.Fatalf("cmdBlocked: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, c.ID) {
		t.Errorf("output should contain blocked issue %s: %q", c.ID, out)
	}
	if !strings.Contains(out, a.ID) {
		t.Errorf("output should list blocker %s: %q", a.ID, out)
	}
	if !strings.Contains(out, b.ID) {
		t.Errorf("output should list blocker %s: %q", b.ID, out)
	}
	if !strings.Contains(out, "Blocker A") {
		t.Errorf("output should contain blocker A title: %q", out)
	}
	if !strings.Contains(out, "Blocker B") {
		t.Errorf("output should contain blocker B title: %q", out)
	}
}

func TestCmdBlockedUnknownFlag(t *testing.T) {
	env := testutil.NewEnv(t)
	defer env.Cleanup()

	var buf bytes.Buffer
	_, err := cmdBlocked(env.Store, []string{"--verbose"}, PlainWriter(&buf), nil)
	if err == nil {
		t.Error("expected error for unknown flag")
	}
}

func TestCmdBlockedJSONEmpty(t *testing.T) {
	env := testutil.NewEnv(t)
	defer env.Cleanup()

	env.Store.Create("No deps", issue.CreateOpts{})
	env.Repo.Commit("create")

	var buf bytes.Buffer
	_, err := cmdBlocked(env.Store, []string{"--json"}, PlainWriter(&buf), nil)
	if err != nil {
		t.Fatalf("cmdBlocked --json: %v", err)
	}

	var result []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 results, got %d", len(result))
	}
}

func TestCmdBlockedPartiallyResolved(t *testing.T) {
	env := testutil.NewEnv(t)
	defer env.Cleanup()

	a, _ := env.Store.Create("Blocker A", issue.CreateOpts{})
	b, _ := env.Store.Create("Blocker B", issue.CreateOpts{})
	c, _ := env.Store.Create("Blocked by two", issue.CreateOpts{})
	env.Store.Link(a.ID, c.ID)
	env.Store.Link(b.ID, c.ID)
	// Close one blocker — issue should still be blocked by the other
	env.Store.Close(a.ID, "done")
	env.Repo.Commit("setup")

	var buf bytes.Buffer
	_, err := cmdBlocked(env.Store, []string{}, PlainWriter(&buf), nil)
	if err != nil {
		t.Fatalf("cmdBlocked: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, c.ID) {
		t.Errorf("issue %s should still be blocked: %q", c.ID, out)
	}
	if !strings.Contains(out, b.ID) {
		t.Errorf("should list remaining blocker %s: %q", b.ID, out)
	}
}

func TestCmdBlockedShowsBlocksLine(t *testing.T) {
	env := testutil.NewEnv(t)
	defer env.Cleanup()

	// Chain: a blocks b blocks c
	a, _ := env.Store.Create("First", issue.CreateOpts{})
	b, _ := env.Store.Create("Middle", issue.CreateOpts{})
	c, _ := env.Store.Create("Last", issue.CreateOpts{})
	env.Store.Link(a.ID, b.ID)
	env.Store.Link(b.ID, c.ID)
	env.Repo.Commit("chain")

	var buf bytes.Buffer
	_, err := cmdBlocked(env.Store, []string{}, PlainWriter(&buf), nil)
	if err != nil {
		t.Fatalf("cmdBlocked: %v", err)
	}
	out := buf.String()

	// b is blocked and also blocks c, so it should show "Blocks: c.ID"
	if !strings.Contains(out, "Blocks: "+c.ID) {
		t.Errorf("blocked output should show Blocks line for middle issue: %q", out)
	}

	if !strings.Contains(out, a.ID) || !strings.Contains(out, "First") {
		t.Errorf("should show b is blocked by a with title: %q", out)
	}
	if !strings.Contains(out, b.ID) || !strings.Contains(out, "Middle") {
		t.Errorf("should show c is blocked by b with title: %q", out)
	}
}

func TestCmdBlockedNoBlocksLineWhenNotBlocking(t *testing.T) {
	env := testutil.NewEnv(t)
	defer env.Cleanup()

	a, _ := env.Store.Create("Blocker", issue.CreateOpts{})
	b, _ := env.Store.Create("Leaf blocked", issue.CreateOpts{})
	env.Store.Link(a.ID, b.ID)
	env.Repo.Commit("link")

	var buf bytes.Buffer
	_, err := cmdBlocked(env.Store, []string{}, PlainWriter(&buf), nil)
	if err != nil {
		t.Fatalf("cmdBlocked: %v", err)
	}
	out := buf.String()

	// b is blocked but doesn't block anything — should not have "Blocks:" line
	if strings.Contains(out, "Blocks:") {
		t.Errorf("leaf blocked issue should not have Blocks line: %q", out)
	}
}

func TestParseBlockedArgs(t *testing.T) {
	ba, err := parseBlockedArgs([]string{})
	if err != nil {
		t.Fatalf("parseBlockedArgs: %v", err)
	}
	if ba.JSON {
		t.Error("expected JSON=false")
	}
}

func TestParseBlockedArgsJSON(t *testing.T) {
	ba, err := parseBlockedArgs([]string{"--json"})
	if err != nil {
		t.Fatalf("parseBlockedArgs: %v", err)
	}
	if !ba.JSON {
		t.Error("expected JSON=true")
	}
}

func TestParseBlockedArgsID(t *testing.T) {
	ba, err := parseBlockedArgs([]string{"bw-abcd"})
	if err != nil {
		t.Fatalf("parseBlockedArgs: %v", err)
	}
	if ba.ID != "bw-abcd" {
		t.Errorf("ID = %q, want bw-abcd", ba.ID)
	}
}

func TestParseBlockedArgsExtraPositional(t *testing.T) {
	_, err := parseBlockedArgs([]string{"bw-abcd", "extra"})
	if err == nil {
		t.Error("expected error for extra positional")
	}
}

func TestCmdBlockedFilterID(t *testing.T) {
	env := testutil.NewEnv(t)
	defer env.Cleanup()

	a, _ := env.Store.Create("Blocker A", issue.CreateOpts{})
	b, _ := env.Store.Create("Blocked B", issue.CreateOpts{})
	c, _ := env.Store.Create("Blocker C", issue.CreateOpts{})
	d, _ := env.Store.Create("Blocked D", issue.CreateOpts{})
	env.Store.Link(a.ID, b.ID)
	env.Store.Link(c.ID, d.ID)
	env.Repo.Commit("setup")

	var buf bytes.Buffer
	_, err := cmdBlocked(env.Store, []string{b.ID}, PlainWriter(&buf), nil)
	if err != nil {
		t.Fatalf("cmdBlocked %s: %v", b.ID, err)
	}
	out := buf.String()
	if !strings.Contains(out, b.ID) || !strings.Contains(out, "Blocked B") {
		t.Errorf("filtered output should contain %s: %q", b.ID, out)
	}
	if !strings.Contains(out, a.ID) || !strings.Contains(out, "Blocker A") {
		t.Errorf("filtered output should contain blocker %s: %q", a.ID, out)
	}
	if strings.Contains(out, d.ID) || strings.Contains(out, "Blocked D") {
		t.Errorf("filtered output should not contain other blocked issue: %q", out)
	}
}

func TestCmdBlockedFilterUnknownID(t *testing.T) {
	env := testutil.NewEnv(t)
	defer env.Cleanup()

	var buf bytes.Buffer
	_, err := cmdBlocked(env.Store, []string{"test-zzzz"}, PlainWriter(&buf), nil)
	if err == nil {
		t.Fatal("expected error for unknown id")
	}
}

func TestCmdBlockedFilterUnblockedID(t *testing.T) {
	env := testutil.NewEnv(t)
	defer env.Cleanup()

	a, _ := env.Store.Create("Free task", issue.CreateOpts{})
	env.Repo.Commit("create")

	var buf bytes.Buffer
	_, err := cmdBlocked(env.Store, []string{a.ID}, PlainWriter(&buf), nil)
	if err != nil {
		t.Fatalf("cmdBlocked: %v", err)
	}
	if !strings.Contains(buf.String(), "no blocked issues") {
		t.Errorf("unblocked issue should report empty graph: %q", buf.String())
	}
}

func TestCmdBlockedNestedChildBySibling(t *testing.T) {
	env := testutil.NewEnv(t)
	defer env.Cleanup()

	epic, _ := env.Store.Create("Epic", issue.CreateOpts{Type: "epic"})
	sib, _ := env.Store.Create("Sibling work", issue.CreateOpts{Parent: epic.ID})
	child, _ := env.Store.Create("Child work", issue.CreateOpts{Parent: epic.ID})
	env.Store.Link(sib.ID, child.ID)
	env.Repo.Commit("setup")

	var buf bytes.Buffer
	_, err := cmdBlocked(env.Store, []string{}, PlainWriter(&buf), nil)
	if err != nil {
		t.Fatalf("cmdBlocked: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, child.ID) || !strings.Contains(out, "Child work") {
		t.Errorf("nested child should appear: %q", out)
	}
	if !strings.Contains(out, sib.ID) || !strings.Contains(out, "Sibling work") {
		t.Errorf("sibling blocker title should appear: %q", out)
	}
}

func TestCmdBlockedClosedBlockerDropsIssue(t *testing.T) {
	env := testutil.NewEnv(t)
	defer env.Cleanup()

	a, _ := env.Store.Create("Blocker", issue.CreateOpts{})
	b, _ := env.Store.Create("Blocked", issue.CreateOpts{})
	env.Store.Link(a.ID, b.ID)
	env.Store.Close(a.ID, "")
	env.Repo.Commit("setup")

	var buf bytes.Buffer
	_, err := cmdBlocked(env.Store, []string{}, PlainWriter(&buf), nil)
	if err != nil {
		t.Fatalf("cmdBlocked: %v", err)
	}
	if !strings.Contains(buf.String(), "no blocked issues") {
		t.Errorf("closed blocker should drop the issue: %q", buf.String())
	}
}
