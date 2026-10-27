package main

import (
	"errors"
	"strings"
	"testing"

	aion2 "github.com/nuriland/aion2-api"
)

func TestItemTools(t *testing.T) {
	ctx := t.Context()

	found, err := searchItems(ctx, tw, searchItemsArgs{Query: "巨劍", Grade: "Epic", Category: "Equip_Weapon", SubCategory: "Greatsword", Size: 2})
	q, _ := fake.last("/dict/search/item")
	if err != nil || len(found.Items) == 0 ||
		q.Get("searchKeyword") != "巨劍" || q.Get("grades") != "Epic" || q.Get("category1") != "Equip_Weapon" || q.Get("category2") != "Greatsword" || q.Get("size") != "2" {
		t.Errorf("search: %v, %d items, sent %v", err, len(found.Items), q)
	}

	if names, err := suggestItems(ctx, tw, suggestItemsArgs{Keyword: "巨"}); err != nil || len(names.Items) == 0 {
		t.Errorf("suggest: %v, %d names", err, len(names.Items))
	}
	if item, err := getItem(ctx, tw, itemArgs{ID: 110120001}); err != nil || item.Name == "" {
		t.Errorf("item: %v, %+v", err, item)
	}
	if grades, err := listItemGrades(ctx, tw, noArgs{}); err != nil || len(grades.Items) == 0 {
		t.Errorf("grades: %v, %d", err, len(grades.Items))
	}
	if tree, err := listItemCategories(ctx, tw, noArgs{}); err != nil || len(tree.Items) == 0 {
		t.Errorf("categories: %v, %d", err, len(tree.Items))
	}

	// Only TW has the catalog, and Global serves no item by id
	if _, err := searchItems(ctx, kr, searchItemsArgs{}); !errors.Is(err, aion2.ErrFeatureUnavailable) {
		t.Errorf("KR search: %v", err)
	}
	if _, err := getItem(ctx, eu, itemArgs{ID: 110120001}); !errors.Is(err, aion2.ErrFeatureUnavailable) {
		t.Errorf("EU item: %v", err)
	}
}

func TestItemFilters(t *testing.T) {
	for _, c := range []struct {
		args searchItemsArgs
		err  string
	}{
		{searchItemsArgs{Query: nothing, Grade: "Epic", Category: "Equip_Weapon", SubCategory: "Greatsword"}, ""},
		{searchItemsArgs{Query: nothing, Grade: "Heroic"}, `no grade "Heroic", it takes Common, Rare, Legend, Unique, Epic`},
		{searchItemsArgs{Query: nothing, Category: "Weapons"}, `no category "Weapons"`},
		{searchItemsArgs{Query: nothing, Category: "Equip_Armor", SubCategory: "Greatsword"}, `no subCategory "Greatsword", it takes Helmet, Shoulder`},
		{searchItemsArgs{Query: nothing, SubCategory: "Shoulder"}, ""},
	} {
		found, err := searchItems(t.Context(), tw, c.args)
		if c.err == "" && (err != nil || found.Items == nil) || c.err != "" && (err == nil || !strings.Contains(err.Error(), c.err)) {
			t.Errorf("%+v: %v, items %v", c.args, err, found.Items)
		}
	}
}
