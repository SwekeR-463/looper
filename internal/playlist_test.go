package internal

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestBuildConcat(t *testing.T) {
	tracks := []Track{
		{Path: "/tmp/01.mp3"},
		{Path: "/tmp/02.mp3"},
		{Path: "/tmp/03.mp3"},
	}

	cases := []struct {
		name   string
		tracks []Track
		gap    string
		want   string
	}{
		{"single track has no gap", tracks[:1], "gap.mp3", "concat:01.mp3"},
		{"gaps sit between tracks", tracks, "gap.mp3", "concat:01.mp3|gap.mp3|02.mp3|gap.mp3|03.mp3"},
		{"empty gap means no gap segments", tracks, "", "concat:01.mp3|02.mp3|03.mp3"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := BuildConcat(tc.tracks, tc.gap); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestReadLinks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "links.txt")
	content := "# my mashup\n\nhttps://a\n  https://b  \n\n# https://skipped\nhttps://c\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ReadLinks(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"https://a", "https://b", "https://c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
