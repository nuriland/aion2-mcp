package main

import (
	"errors"
	"testing"

	aion2 "github.com/nuriland/aion2-api"
)

func TestRankings(t *testing.T) {
	ctx := t.Context()

	board, err := getRankings(ctx, eu, rankingArgs{Contents: "arena_cooperation", ServerID: 1001})
	if err != nil {
		t.Fatal(err)
	}
	if q, _ := fake.last("/ranking/list"); board.Season == nil || len(board.Entries) == 0 || q.Get("rankingContentsType") != "6" || q.Get("serverId") != "1001" {
		t.Errorf("board: season %v, %d entries, sent %v", board.Season, len(board.Entries), q)
	}

	// What every board answers while NC has them off
	if _, err := getRankings(ctx, eu, rankingArgs{Contents: "abyss", ServerID: 2001}); !errors.Is(err, aion2.ErrNoSeason) {
		t.Errorf("no season: %v", err)
	}
}
