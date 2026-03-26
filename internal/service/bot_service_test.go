package service

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractFirstURL_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want string
	}{
		{name: "empty text", text: "", want: ""},
		{name: "without url", text: "just words", want: ""},
		{name: "single https", text: "check https://example.com/video", want: "https://example.com/video"},
		{name: "single http", text: "http://example.com", want: "http://example.com"},
		{name: "first url from many", text: "a https://one.test b https://two.test", want: "https://one.test"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &BotService{urlRegexp: regexp.MustCompile(`https?://\S+`)}
			got := svc.extractFirstURL(tt.text)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractComment_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want string
	}{
		{name: "empty text", text: "", want: ""},
		{name: "single line only", text: "https://example.com", want: ""},
		{name: "comment on second line", text: "https://example.com\nmy comment", want: "my comment"},
		{name: "trim comment", text: "https://example.com\n   spaced comment   ", want: "spaced comment"},
		{name: "multiline comment", text: "https://example.com\nline1\nline2", want: "line1\nline2"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := extractComment(tt.text)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTooLarge_TableDriven(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	smallPath := filepath.Join(tmp, "small.mp4")
	largePath := filepath.Join(tmp, "large.mp4")
	missingPath := filepath.Join(tmp, "missing.mp4")

	createSizedFile(t, smallPath, maxTelegramVideoSize-1)
	createSizedFile(t, largePath, maxTelegramVideoSize+1)

	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "missing file", path: missingPath, want: false},
		{name: "smaller than limit", path: smallPath, want: false},
		{name: "greater than limit", path: largePath, want: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := tooLarge(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func createSizedFile(t *testing.T, path string, size int64) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}
	t.Cleanup(func() {
		_ = f.Close()
	})

	if err := f.Truncate(size); err != nil {
		t.Fatalf("truncate file: %v", err)
	}
}
