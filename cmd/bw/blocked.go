package main

import (
	"fmt"
	"strings"

	"github.com/bkono/beadwork/internal/config"

	"github.com/bkono/beadwork/internal/issue"
	"github.com/bkono/beadwork/internal/md"
)

type BlockedArgs struct {
	ID   string
	JSON bool
}

func parseBlockedArgs(raw []string) (BlockedArgs, error) {
	a, err := ParseArgs(raw, nil, []string{"--json"})
	if err != nil {
		return BlockedArgs{}, err
	}
	if len(a.Pos()) > 1 {
		return BlockedArgs{}, fmt.Errorf("usage: bw blocked [id] [--json]")
	}
	return BlockedArgs{ID: a.PosFirst(), JSON: a.JSON()}, nil
}

func cmdBlocked(store *issue.Store, args []string, w Writer, _ *config.Config) (*config.Config, error) {
	ba, err := parseBlockedArgs(args)
	if err != nil {
		return nil, err
	}

	var filterID string
	if ba.ID != "" {
		iss, err := store.Get(ba.ID)
		if err != nil {
			return nil, err
		}
		filterID = iss.ID
	}

	blocked, err := store.Blocked()
	if err != nil {
		return nil, err
	}

	if filterID != "" {
		var filtered []issue.BlockedIssue
		for _, bi := range blocked {
			if bi.ID == filterID {
				filtered = append(filtered, bi)
			}
		}
		blocked = filtered
	}

	if ba.JSON {
		fprintJSON(w, blocked)
		return nil, nil
	}

	if len(blocked) == 0 {
		fmt.Fprintln(w, "no blocked issues")
		return nil, nil
	}

	fmt.Fprintf(w, "\nBlocked (%d):\n", len(blocked))

	for _, bi := range blocked {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "{p:%d} {id:%s} %s  %s\n",
			bi.Priority, bi.ID, bi.Status, md.Escape(bi.Title))
		w.Push(2)

		fmt.Fprintln(w, "Blocked by:")
		for _, blocker := range bi.OpenBlockers {
			if blocker.Title == "" {
				fmt.Fprintf(w, "{id:%s} %s\n", blocker.ID, blocker.Status)
			} else {
				fmt.Fprintf(w, "{id:%s} %s  %s\n",
					blocker.ID, blocker.Status, md.Escape(blocker.Title))
			}
		}

		if len(bi.Blocks) > 0 {
			blockIDs := make([]string, len(bi.Blocks))
			copy(blockIDs, bi.Blocks)
			fmt.Fprintf(w, "Blocks: %s\n", strings.Join(blockIDs, ", "))
		}
		w.Pop()
	}
	return nil, nil
}
