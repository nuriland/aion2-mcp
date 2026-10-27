package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	aion2 "github.com/nuriland/aion2-api"
)

// noGear is a made-up characterId the fake NC serves a profile for, but refuses equipment
const noGear = "no-gear"

var searches = map[string]func(page int) ([]string, int){
	"many": func(page int) ([]string, int) {
		return []string{fmt.Sprintf("many%da", page), fmt.Sprintf("many%db", page)}, 9
	},
	"nobody": func(int) ([]string, int) { return nil, 0 },
	"naked":  func(int) ([]string, int) { return []string{"Naked"}, 1 },
	"deep": func(page int) ([]string, int) {
		if page == 3 {
			return []string{"Deeper3", "Deep"}, 4
		}
		return []string{fmt.Sprintf("Deeper%d", page)}, 4
	},
}

func searchPage(w http.ResponseWriter, q url.Values) {
	page, _ := strconv.Atoi(q.Get("page"))
	names, last := searches[q.Get("keyword")](page)

	type row struct {
		CharacterID string `json:"characterId"`
		Name        string `json:"name"`
		Race        int    `json:"race"`
		PcID        int    `json:"pcId"`
		Level       int    `json:"level"`
		ServerID    int    `json:"serverId"`
		ServerName  string `json:"serverName"`
	}
	rows := []row{}
	for _, name := range names {
		id := "id-" + name
		if name == "Naked" {
			id = noGear
		}
		rows = append(rows, row{CharacterID: id, Name: name, Race: 1, PcID: 29, Level: 50, ServerID: 1001, ServerName: "시엘"})
	}
	json.NewEncoder(w).Encode(map[string]any{
		"list":       rows,
		"pagination": map[string]int{"page": page, "size": searchSize, "total": last * len(names), "endPage": last},
	})
}

// The fake NC is shared by every test, so each case counts the searches it made from where the last one left off
func TestFindCharacterPages(t *testing.T) {
	ctx := t.Context()

	// Across servers a second "Deep" could sit on page 4, so every page of both races is read
	_, before := fake.last("/character")
	got, err := findCharacter(ctx, tw, findArgs{Name: "deep"})
	if _, n := fake.last("/character"); err != nil || got.Character == nil || n-before != 8 {
		t.Errorf("deep on every server: %v, %d searches, %+v", err, n-before, got)
	}

	// On one server the first match is the only one
	_, before = fake.last("/character")
	got, err = findCharacter(ctx, tw, findArgs{Name: "deep", Server: "1001"})
	if _, n := fake.last("/character"); err != nil || got.Character == nil || n-before != 3 {
		t.Errorf("deep on one server: %v, %d searches, %+v", err, n-before, got)
	}

	_, before = fake.last("/character")
	got, err = findCharacter(ctx, tw, findArgs{Name: "many"})
	if _, n := fake.last("/character"); err != nil || got.Character != nil || len(got.Candidates) != maxCandidates || !strings.Contains(got.Note, "there are more") || n-before != 2*maxPages {
		t.Errorf("many: %v, %d searches, %+v", err, n-before, got)
	}

	// Global takes both races in one search
	_, before = fake.last("/search/v2/character")
	_, err = findCharacter(ctx, sa, findArgs{Name: "many"})
	if q, n := fake.last("/search/v2/character"); err != nil || n-before != maxPages || q.Has("race") {
		t.Errorf("Global: %v, %d searches, last %v", err, n-before, q)
	}
}

func TestFindCharacterMisses(t *testing.T) {
	ctx := t.Context()

	got, err := findCharacter(ctx, tw, findArgs{Name: "nobody"})
	if err != nil || got.Character != nil || len(got.Candidates) != 0 || !strings.Contains(got.Note, "contains") {
		t.Errorf("nobody: %v, %+v", err, got)
	}

	got, err = findCharacter(ctx, tw, findArgs{Name: "naked"})
	if err != nil || got.Character == nil || got.Equipment != nil || !strings.Contains(got.Note, "equipment unavailable") {
		t.Errorf("missing gear should keep the profile: %v, %+v", err, got)
	}
}

func TestFindCharacter(t *testing.T) {
	ctx := t.Context()

	_, before := fake.last("/search/character")
	got, err := findCharacter(ctx, tw, findArgs{Name: "  ALPHA ", Server: "시엘"})
	if err != nil || got.Character == nil || got.Character.Profile.Name != "Alpha" || got.Equipment == nil || len(got.Candidates) != 0 {
		t.Errorf("exact match on a named server: %v, %+v", err, got)
	}
	if q, n := fake.last("/search/character"); n-before != 1 || q.Get("serverId") != "1001" || q.Get("race") != "1" {
		t.Errorf("a named server should mean one search, on its id and race: %d searches, last %v", n-before, q)
	}

	// Without a server both races are searched
	_, before = fake.last("/search/character")
	got, err = findCharacter(ctx, tw, findArgs{Name: "beta"})
	if _, n := fake.last("/search/character"); err != nil || got.Character == nil || n-before != 2 {
		t.Errorf("exact match across races: %v, %d searches, %+v", err, n-before, got)
	}

	got, err = findCharacter(ctx, tw, findArgs{Name: "a"})
	if err != nil || got.Character != nil || len(got.Candidates) != 2 || got.Note == "" {
		t.Errorf("no exact match should list the near misses: %v, %+v", err, got)
	}

	if _, err := findCharacter(ctx, tw, findArgs{Name: "Alpha", Server: "Israphel"}); !errors.Is(err, aion2.ErrNotFound) || !strings.Contains(err.Error(), "servers are 시엘") {
		t.Errorf("unknown server: %v", err)
	}
}
