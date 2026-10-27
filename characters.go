package main

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	aion2 "github.com/nuriland/aion2-api"
)

func characterTools(s *mcp.Server, cs *clients) {
	add(s, cs, "find_character", "Look a character up by exact name, optionally on one server, and return its profile, stats and equipment. "+
		"With no exact match, or several, returns candidates instead", findCharacter)
	add(s, cs, "search_characters", "Search characters by name. KR and TW search one race at a time, so they need raceId. "+
		"Each hit's ref opens get_character, get_equipment and get_daevanion", searchCharacters)
	add(s, cs, "get_character", "A character's profile, stats, titles, rankings and Daevanion board progress", getCharacter)
	add(s, cs, "get_equipment", "A character's worn gear, skins, pet, wings and skills. A slot opens get_equipped_item", getEquipment)
	add(s, cs, "get_equipped_item", "One worn item in full, at the enchant level get_equipment shows: its stats, random rolls, magic stones and god stones", getEquippedItem)
	add(s, cs, "get_daevanion", "One of a character's Daevanion boards: every node, open or not, and the effects the open ones give", getDaevanion)
}

// characterArgs is a CharacterRef, as search_characters hands them out
type characterArgs struct {
	scope

	ServerID    int    `json:"serverId" jsonschema:"the ref's serverId"`
	CharacterID string `json:"characterId" jsonschema:"the ref's characterId, or the id from a profile URL"`
}

func (args characterArgs) ref() aion2.CharacterRef {
	return aion2.CharacterRef{ServerID: args.ServerID, CharacterID: args.CharacterID}
}

type searchCharactersArgs struct {
	scope

	Keyword  string     `json:"keyword" jsonschema:"part of the character's name"`
	RaceID   int        `json:"raceId,omitempty" jsonschema:"1 Elyos, 2 Asmodian, KR and TW need one, Global may leave it out"`
	ServerID int        `json:"serverId,omitempty" jsonschema:"one server, empty is every server in the region"`
	ClassIDs []int      `json:"classIds,omitempty" jsonschema:"class ids from list_classes"`
	Page     pageNumber `json:"page,omitempty" jsonschema:"from 1"`
	Size     pageSize   `json:"size,omitempty" jsonschema:"default 40, up to 200"`
}

// equippedItemArgs is a slot from get_equipment
type equippedItemArgs struct {
	characterArgs

	ItemID  int `json:"itemId" jsonschema:"the slot's id from get_equipment"`
	SlotPos int `json:"slotPos" jsonschema:"the slot's slotPos from get_equipment"`
}

type daevanionArgs struct {
	characterArgs

	BoardID int `json:"boardId" jsonschema:"a daevanion id from get_character"`
}

func searchCharacters(ctx context.Context, c *aion2.Client, args searchCharactersArgs) (page[aion2.CharacterSummary], error) {
	p, err := c.SearchCharacters(ctx, aion2.CharacterSearch{
		Keyword:  args.Keyword,
		RaceID:   args.RaceID,
		ServerID: args.ServerID,
		ClassIDs: args.ClassIDs,
		Page:     int(args.Page),
		Size:     int(args.Size),
	})
	if err != nil {
		return page[aion2.CharacterSummary]{}, err
	}
	return pageOf(p), nil
}

func getCharacter(ctx context.Context, c *aion2.Client, args characterArgs) (*aion2.Character, error) {
	character, err := c.Character(ctx, args.ref())
	return character, args.blame(ctx, c, err)
}

func getEquipment(ctx context.Context, c *aion2.Client, args characterArgs) (*aion2.Equipment, error) {
	equipment, err := c.Equipment(ctx, args.ref())
	return equipment, args.blame(ctx, c, err)
}

// getEquippedItem sends the slot as get_equipment has it, since NC works the stats out at whatever enchantLevel it is sent
func getEquippedItem(ctx context.Context, c *aion2.Client, args equippedItemArgs) (*aion2.EquippedItem, error) {
	equipment, err := c.Equipment(ctx, args.ref())
	if err != nil {
		return nil, args.blame(ctx, c, err)
	}
	for _, slot := range slices.Concat(equipment.Slots, equipment.Skins) {
		if slot.ItemID == args.ItemID && slot.SlotPos == args.SlotPos {
			return c.EquippedItem(ctx, args.ref(), slot)
		}
	}
	return nil, fmt.Errorf("%w: item %d is not worn in slot %d, get_equipment has what is", aion2.ErrNotFound, args.ItemID, args.SlotPos)
}

func getDaevanion(ctx context.Context, c *aion2.Client, args daevanionArgs) (*aion2.DaevanionBoard, error) {
	board, err := c.Daevanion(ctx, args.ref(), args.BoardID)
	return board, args.blame(ctx, c, err)
}

// blame says when it is the server that is unknown, as NC answers an unknown server and an unknown character with the same not found
func (args characterArgs) blame(ctx context.Context, c *aion2.Client, err error) error {
	if !errors.Is(err, aion2.ErrNotFound) {
		return err
	}
	servers, serr := c.Servers(ctx)
	if serr != nil || slices.ContainsFunc(servers, func(s aion2.Server) bool { return s.ServerID == args.ServerID }) {
		return err
	}
	return fmt.Errorf("%w: no server %d on %s, list_servers has them", aion2.ErrNotFound, args.ServerID, c.Region())
}
