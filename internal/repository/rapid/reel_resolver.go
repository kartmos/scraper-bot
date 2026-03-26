package rapid

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	host     = "instagram-scraper-api2.p.rapidapi.com"
	endpoint = "https://instagram-scraper-api2.p.rapidapi.com/v1/post_info?code_or_id_or_url="
)

type ReelResolver struct {
	client      *http.Client
	tokens      []string
	videoRegexp *regexp.Regexp
}

func NewReelResolver(tokens []string, timeout time.Duration) *ReelResolver {
	return &ReelResolver{
		client:      &http.Client{Timeout: timeout},
		tokens:      tokens,
		videoRegexp: regexp.MustCompile(`"video_url":"([^"]+)"`),
	}
}

func (r *ReelResolver) ResolveVideoURL(ctx context.Context, reelLink string) (string, error) {
	if len(r.tokens) == 0 {
		return "", fmt.Errorf("rapid tokens are empty")
	}

	requestURL := endpoint + url.QueryEscape(reelLink)

	for _, token := range r.tokens {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
		if err != nil {
			continue
		}

		req.Header.Set("x-rapidapi-key", token)
		req.Header.Set("x-rapidapi-host", host)

		res, err := r.client.Do(req)
		if err != nil {
			continue
		}

		body, readErr := io.ReadAll(res.Body)
		if closeErr := res.Body.Close(); closeErr != nil && readErr == nil {
			readErr = closeErr
		}
		if readErr != nil {
			continue
		}

		if res.StatusCode < 200 || res.StatusCode >= 300 {
			continue
		}

		match := r.videoRegexp.FindStringSubmatch(string(body))
		if len(match) == 2 && strings.TrimSpace(match[1]) != "" {
			return match[1], nil
		}
	}

	return "", fmt.Errorf("failed to resolve reel url")
}
