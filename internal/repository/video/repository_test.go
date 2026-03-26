package video

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type resolverStub struct {
	resolve func(ctx context.Context, reelLink string) (string, error)
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func (r resolverStub) ResolveVideoURL(ctx context.Context, reelLink string) (string, error) {
	return r.resolve(ctx, reelLink)
}

func TestRepositoryDownload_TableDriven(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	ytDLPPath := createFakeYTDLP(t, tmp)

	tests := []struct {
		name      string
		sourceURL string
		repo      *Repository
		wantErr   string
	}{
		{
			name:      "youtube via yt-dlp",
			sourceURL: "https://youtube.com/watch?v=abc",
			repo:      NewRepository(filepath.Join(tmp, "yt"), ytDLPPath, nil),
		},
		{
			name:      "instagram via resolver and direct download",
			sourceURL: "https://www.instagram.com/reel/abc/",
			repo: NewRepository(filepath.Join(tmp, "ig"), ytDLPPath, resolverStub{resolve: func(_ context.Context, _ string) (string, error) {
				return "https://cdn.test/video.mp4", nil
			}}),
		},
		{
			name:      "instagram without resolver",
			sourceURL: "https://instagram.com/reel/abc/",
			repo:      NewRepository(filepath.Join(tmp, "ig-no-resolver"), ytDLPPath, nil),
			wantErr:   "reel resolver is not configured",
		},
		{
			name:      "unsupported host",
			sourceURL: "https://example.org/video",
			repo:      NewRepository(filepath.Join(tmp, "unsupported"), ytDLPPath, nil),
			wantErr:   "unsupported host",
		},
		{
			name:      "invalid link",
			sourceURL: "not-a-link",
			repo:      NewRepository(filepath.Join(tmp, "invalid"), ytDLPPath, nil),
			wantErr:   "invalid link",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if strings.Contains(tt.name, "instagram via resolver") {
				tt.repo.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader("video-bytes")),
						Header:     make(http.Header),
					}, nil
				})
			}

			path, err := tt.repo.Download(context.Background(), 42, tt.sourceURL)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Empty(t, path)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, filepath.Join(tt.repo.downloadDir, "42.mp4"), path)

			content, readErr := os.ReadFile(path)
			require.NoError(t, readErr)
			assert.NotEmpty(t, content)
		})
	}
}

func TestDownloadByURL_Non2xx(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	repo := NewRepository(tmp, "yt-dlp", nil)
	repo.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Status:     "502 Bad Gateway",
			Body:       io.NopCloser(strings.NewReader("upstream failed")),
			Header:     make(http.Header),
		}, nil
	})

	err := repo.downloadByURL(context.Background(), "https://cdn.test/video.mp4", filepath.Join(tmp, "out.mp4"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "download file failed with status")
}

func TestExtractHost_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rawURL  string
		want    string
		wantErr bool
	}{
		{name: "valid host", rawURL: "https://www.youtube.com/watch?v=1", want: "www.youtube.com"},
		{name: "valid host without path", rawURL: "https://youtu.be", want: "youtu.be"},
		{name: "invalid", rawURL: "not-url", wantErr: true},
		{name: "missing host", rawURL: "/relative/path", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			host, err := extractHost(tt.rawURL)
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, host)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, host)
		})
	}
}

func createFakeYTDLP(t *testing.T, dir string) string {
	t.Helper()

	path := filepath.Join(dir, "yt-dlp")
	script := `#!/bin/sh
set -eu
out=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "-o" ]; then
    shift
    out="$1"
  fi
  shift || true
done
if [ -z "$out" ]; then
  echo "missing output" >&2
  exit 1
fi
printf '%s' 'yt-dlp-video' > "$out"
`

	require.NoError(t, os.WriteFile(path, []byte(script), 0o755))

	deadline := time.Now().Add(2 * time.Second)
	for {
		if info, err := os.Stat(path); err == nil && info.Mode()&0o111 != 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("fake yt-dlp is not executable: %s", path)
		}
		time.Sleep(10 * time.Millisecond)
	}

	return path
}

func TestRepositoryDownload_ResolverError(t *testing.T) {
	t.Parallel()

	repo := NewRepository(t.TempDir(), "yt-dlp", resolverStub{resolve: func(_ context.Context, _ string) (string, error) {
		return "", fmt.Errorf("resolver failed")
	}})

	_, err := repo.Download(context.Background(), 1, "https://instagram.com/reel/abc/")
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "resolver failed"))
}
