package video

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type ReelResolver interface {
	ResolveVideoURL(ctx context.Context, reelLink string) (string, error)
}

type Repository struct {
	downloadDir  string
	ytDLPPath    string
	reelResolver ReelResolver
	httpClient   *http.Client
}

func NewRepository(downloadDir, ytDLPPath string, reelResolver ReelResolver) *Repository {
	return &Repository{
		downloadDir:  downloadDir,
		ytDLPPath:    ytDLPPath,
		reelResolver: reelResolver,
		httpClient:   &http.Client{},
	}
}

func (r *Repository) Download(ctx context.Context, updateID int, sourceLink string) (string, error) {
	if err := os.MkdirAll(r.downloadDir, 0o755); err != nil {
		return "", fmt.Errorf("create download dir: %w", err)
	}

	host, err := extractHost(sourceLink)
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(r.downloadDir, strconv.Itoa(updateID)+".mp4")

	switch host {
	case "vt.tiktok.com", "youtube.com", "www.youtube.com", "youtu.be":
		if err := r.downloadWithYTDLP(ctx, sourceLink, filePath); err != nil {
			return "", err
		}
	case "www.instagram.com", "instagram.com":
		if r.reelResolver == nil {
			return "", fmt.Errorf("reel resolver is not configured")
		}
		mediaURL, err := r.reelResolver.ResolveVideoURL(ctx, sourceLink)
		if err != nil {
			return "", err
		}
		if err := r.downloadByURL(ctx, mediaURL, filePath); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported host: %s", host)
	}

	return filePath, nil
}

func (r *Repository) downloadWithYTDLP(ctx context.Context, link, outputPath string) error {
	cmd := exec.CommandContext(ctx, r.ytDLPPath, "-f", "mp4", "-o", outputPath, link)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("yt-dlp download error: %w (%s)", err, string(out))
	}
	return nil
}

func (r *Repository) downloadByURL(ctx context.Context, fileURL, outputPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	res, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download file: %w", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("download file failed with status: %s", res.Status)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	if _, err := io.Copy(file, res.Body); err != nil {
		return fmt.Errorf("copy downloaded body: %w", err)
	}

	return nil
}

func extractHost(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return "", fmt.Errorf("invalid link")
	}
	return parsed.Host, nil
}
