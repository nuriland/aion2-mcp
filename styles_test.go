package main

import (
	"slices"
	"testing"
)

func TestStyleTools(t *testing.T) {
	ctx := t.Context()

	top, err := topStyles(ctx, eu, topStylesArgs{SortBy: "LIKES", Period: "DAY_30", Gender: "FEMALE"})
	if q, _ := fake.last("/top100/aion2_global/"); err != nil || len(top.Items) == 0 || q.Get("sortBy") != "LIKES" || q.Get("period") != "DAY_30" || q.Get("charGender") != "FEMALE" {
		t.Errorf("top: %v, %d styles, sent %v", err, len(top.Items), q)
	}

	found, err := searchStyles(ctx, eu, searchStylesArgs{Keyword: "x", Field: "item", Tags: []string{"LOVELY", "CUTE"}, Class: "GLADIATOR", Page: 2, Size: 2})
	q, _ := fake.last("/search/aion2_global/")
	if err != nil || len(found.Items) == 0 || q.Get("keyword") != "x" || q.Get("field") != "item" || q.Get("charClass") != "GLADIATOR" ||
		!slices.Equal(q["userTags"], []string{"LOVELY", "CUTE"}) || q.Get("page") != "1" || q.Get("size") != "2" { // NC counts pages from 0
		t.Errorf("search: %v, %d styles, sent %v", err, len(found.Items), q)
	}

	if tags, err := listStyleTags(ctx, eu, noArgs{}); err != nil || len(tags.Items) != 6 || tags.Items[0].Code != "LOVELY" || tags.Items[2].Type != "CLASS" {
		t.Errorf("tags: %v, %+v", err, tags.Items)
	}
	if style, err := getStyle(ctx, eu, styleArgs{ID: "x"}); err != nil || style.Title == "" {
		t.Errorf("style: %v, %+v", err, style)
	}
	if replies, err := listStyleComments(ctx, eu, styleCommentsArgs{styleArgs: styleArgs{ID: "x"}, Limit: 1}); err != nil || len(replies.Items) != 1 || !replies.Truncated {
		t.Errorf("limit 1: %v, %d comments, truncated %v", err, len(replies.Items), replies.Truncated)
	}
}
