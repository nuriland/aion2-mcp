package main

import "testing"

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

	item, err := getEquippedItem(ctx, eu, equippedItemArgs{characterArgs: ref, ItemID: 110720001, SlotPos: 1, EnchantLevel: 20})
	if err != nil {
		t.Fatal(err)
	}
	q, _ := fake.last("/character/equipment/item")
	if item.Name == "" || item.SlotPos != 1 || q.Get("id") != "110720001" || q.Get("slotPos") != "1" || q.Get("enchantLevel") != "20" {
		t.Errorf("equipped item %q in slot %d, sent %v", item.Name, item.SlotPos, q)
	}

	board, err := getDaevanion(ctx, eu, daevanionArgs{characterArgs: ref, BoardID: 71})
	if err != nil {
		t.Fatal(err)
	}
	if q, _ := fake.last("/character/daevanion/detail"); board.BoardID != 71 || len(board.Nodes) == 0 || q.Get("boardId") != "71" {
		t.Errorf("daevanion board %d, %d nodes, sent %v", board.BoardID, len(board.Nodes), q)
	}
}
