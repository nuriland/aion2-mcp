package main

import (
	"errors"
	"strings"
	"testing"

	aion2 "github.com/nuriland/aion2-api"
)

func TestCharacterTools(t *testing.T) {
	ctx := t.Context()
	ref := characterArgs{ServerID: 1001, CharacterID: "abc="}

	ch, err := getCharacter(ctx, eu, ref)
	if err != nil {
		t.Fatal(err)
	}
	if q, _ := fake.last("/character/info"); ch.Profile.Name != "Alpha" || len(ch.Stats) == 0 || q.Get("serverId") != "1001" || q.Get("characterId") != "abc=" {
		t.Errorf("character %+v, sent %v", ch.Profile, q)
	}

	if eq, err := getEquipment(ctx, eu, ref); err != nil || len(eq.Slots) == 0 {
		t.Errorf("equipment: %v, %+v", err, eq)
	}

	item, err := getEquippedItem(ctx, eu, equippedItemArgs{characterArgs: ref, ItemID: 110720001, SlotPos: 1})
	if err != nil {
		t.Fatal(err)
	}
	q, _ := fake.last("/character/equipment/item")
	if item.Name == "" || item.SlotPos != 1 || q.Get("id") != "110720001" || q.Get("slotPos") != "1" || q.Get("enchantLevel") != "20" {
		t.Errorf("equipped item %q in slot %d, sent %v", item.Name, item.SlotPos, q)
	}
	if _, err := getEquippedItem(ctx, eu, equippedItemArgs{characterArgs: ref, ItemID: 110720001, SlotPos: 2}); !errors.Is(err, aion2.ErrNotFound) {
		t.Errorf("an item in a slot it is not worn in: %v", err)
	}

	board, err := getDaevanion(ctx, eu, daevanionArgs{characterArgs: ref, BoardID: 71})
	if err != nil {
		t.Fatal(err)
	}
	if q, _ := fake.last("/character/daevanion/detail"); board.BoardID != 71 || len(board.Nodes) == 0 || q.Get("boardId") != "71" {
		t.Errorf("daevanion board %d, %d nodes, sent %v", board.BoardID, len(board.Nodes), q)
	}
}

func TestUnknownServer(t *testing.T) {
	for _, c := range []struct {
		serverID int
		blamed   bool
	}{
		{1001, false},
		{9999, true},
	} {
		_, err := getCharacter(t.Context(), eu, characterArgs{ServerID: c.serverID, CharacterID: missing})
		if !errors.Is(err, aion2.ErrNotFound) || strings.Contains(err.Error(), "no server 9999") != c.blamed {
			t.Errorf("server %d: %v", c.serverID, err)
		}
	}
}
