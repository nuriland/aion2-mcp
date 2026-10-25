package main

import (
	"strings"
	"testing"

	aion2 "github.com/nuriland/aion2-api"
)

func TestListPostsLimit(t *testing.T) {
	ctx := t.Context()

	all, err := listPosts(ctx, eu, postsArgs{boardArgs: boardArgs{Board: aion2.BoardPatchNotes}, Limit: 200})
	if err != nil || len(all.Items) < 2 || all.Truncated {
		t.Fatalf("%v: %d posts, truncated %v", err, len(all.Items), all.Truncated)
	}

	one, err := listPosts(ctx, eu, postsArgs{boardArgs: boardArgs{Board: aion2.BoardPatchNotes}, Limit: 1})
	if err != nil || len(one.Items) != 1 || !one.Truncated || one.Items[0].ID != all.Items[0].ID {
		t.Errorf("limit 1: %v, %d posts, truncated %v", err, len(one.Items), one.Truncated)
	}
}

func TestGetPost(t *testing.T) {
	ctx := t.Context()
	args := postArgs{boardArgs: boardArgs{Board: aion2.BoardPatchNotes}, ID: "6ab2d738a279104f7d9d5d5b"}

	plain, err := getPost(ctx, eu, args)
	if err != nil {
		t.Fatal(err)
	}
	if plain.HTML != "" || plain.Title == "" || strings.Contains(plain.Text, "<") {
		t.Errorf("text format: title %q, html %d bytes, text %q", plain.Title, len(plain.HTML), plain.Text)
	}

	args.Format = "html"
	raw, err := getPost(ctx, eu, args)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(raw.HTML, "<") || raw.Text != "" {
		t.Errorf("html format: html %d bytes, text %q", len(raw.HTML), raw.Text)
	}
}

func TestPinnedPostsAndComments(t *testing.T) {
	ctx := t.Context()

	if pinned, err := listPinnedPosts(ctx, eu, boardArgs{Board: aion2.BoardNotices}); err != nil || len(pinned.Items) == 0 {
		t.Errorf("pinned: %v, %d posts", err, len(pinned.Items))
	}

	args := commentsArgs{boardArgs: boardArgs{Board: aion2.BoardFree}, PostID: "x", Limit: 200}
	all, err := listComments(ctx, eu, args)
	if err != nil {
		t.Fatal(err)
	}
	args.Limit = 1
	one, err := listComments(ctx, eu, args)
	if err != nil || len(all.Items) < 2 || all.Truncated || len(one.Items) != 1 || !one.Truncated || one.Items[0].ID != all.Items[0].ID {
		t.Errorf("all %d (truncated %v), limit 1: %v, %d (truncated %v)", len(all.Items), all.Truncated, err, len(one.Items), one.Truncated)
	}
}
