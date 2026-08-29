package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/bkono/beadwork/internal/config"

	"github.com/bkono/beadwork/internal/issue"
	"github.com/bkono/beadwork/internal/md"
	"github.com/bkono/beadwork/internal/repo"
	"github.com/bkono/beadwork/internal/tmpl"
	"github.com/bkono/beadwork/prompts"
)

type PrimeData struct {
	Prefix           string
	WorktreeDirty    bool
	Git              repo.GitContext
	OverdueCount     int
	ExpiredDeferrals string
}

func cmdPrime(store *issue.Store, _ []string, w Writer, _ *config.Config) (*config.Config, error) {
	r := store.Committer.(*repo.Repo)
	cfg := r.ListConfig()
	gitCtx := r.GetGitContext()

	overdueIssues, _ := store.List(issue.Filter{Overdue: true})

	// Find expired deferrals for the reminders section.
	now := store.Now()
	deferredIssues, _ := store.List(issue.Filter{Status: "deferred"})
	var expiredLines []string
	for _, iss := range deferredIssues {
		if issue.IsDeferralExpired(iss.DeferUntil, now) {
			expiredLines = append(expiredLines, md.IssueOneLinerWithDue(iss, now, nil))
		}
	}

	data := PrimeData{
		Prefix:           cfg["prefix"],
		WorktreeDirty:    gitCtx.Dirty,
		Git:              gitCtx,
		OverdueCount:     len(overdueIssues),
		ExpiredDeferrals: strings.Join(expiredLines, "\n"),
	}

	bwFn := func(args ...string) string {
		if cmd := commandMap[args[0]]; cmd != nil {
			var buf bytes.Buffer
			cmd.Run(store, args[1:], TokenWriter(&buf), nil)
			return strings.TrimRight(buf.String(), "\n")
		}
		return ""
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, "prime", prompts.Prime, data, bwFn); err != nil {
		return nil, err
	}

	out := strings.Trim(buf.String(), "\n")
	fmt.Fprint(w, out)
	fmt.Fprintln(w)
	return nil, nil
}
