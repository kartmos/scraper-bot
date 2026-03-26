package rapid

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestResolveVideoURL_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		tokens    []string
		transport roundTripFunc
		wantURL   string
		wantErr   string
	}{
		{
			name:    "empty tokens",
			tokens:  nil,
			wantErr: "rapid tokens are empty",
		},
		{
			name:   "fallback to second token",
			tokens: []string{"token-1", "token-2"},
			transport: func(req *http.Request) (*http.Response, error) {
				switch req.Header.Get("x-rapidapi-key") {
				case "token-1":
					return &http.Response{StatusCode: http.StatusTooManyRequests, Body: io.NopCloser(strings.NewReader("rate limit"))}, nil
				case "token-2":
					return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"video_url":"https://cdn.test/reel.mp4"}`))}, nil
				default:
					return nil, errors.New("unexpected token")
				}
			},
			wantURL: "https://cdn.test/reel.mp4",
		},
		{
			name:   "all attempts fail",
			tokens: []string{"token-1", "token-2"},
			transport: func(req *http.Request) (*http.Response, error) {
				if req.Header.Get("x-rapidapi-key") == "token-1" {
					return nil, errors.New("network error")
				}
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"no_video":true}`))}, nil
			},
			wantErr: "failed to resolve reel url",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewReelResolver(tt.tokens, time.Second)
			if tt.transport != nil {
				resolver.client.Transport = tt.transport
			}

			got, err := resolver.ResolveVideoURL(context.Background(), "https://instagram.com/reel/abc/")
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Empty(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantURL, got)
		})
	}
}

func TestResolveVideoURL_SetsHeadersAndEscapesURL(t *testing.T) {
	t.Parallel()

	var capturedHost string
	var capturedToken string
	var capturedURL string

	resolver := NewReelResolver([]string{"token-1"}, time.Second)
	resolver.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		capturedHost = req.Header.Get("x-rapidapi-host")
		capturedToken = req.Header.Get("x-rapidapi-key")
		capturedURL = req.URL.String()
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"video_url":"https://cdn.test/reel.mp4"}`))}, nil
	})

	_, err := resolver.ResolveVideoURL(context.Background(), "https://instagram.com/reel/test/?a=1&b=2")
	require.NoError(t, err)

	assert.Equal(t, host, capturedHost)
	assert.Equal(t, "token-1", capturedToken)
	assert.Contains(t, capturedURL, "code_or_id_or_url=https%3A%2F%2Finstagram.com%2Freel%2Ftest%2F%3Fa%3D1%26b%3D2")
}
