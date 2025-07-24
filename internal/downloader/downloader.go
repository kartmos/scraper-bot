// cSpell:disable
package downloader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	cfg "github.com/kartmos/scraper-bot/config"
	rapid "github.com/kartmos/scraper-bot/internal/external/rapid"
)

const (
	PatInsta      = `https://www.instagram.com/reel/([^"]+)/`
	videoPatInsta = `"video_url":"([^"]+)",`
	apiURL1       = "https://instagram-scraper-api2.p.rapidapi.com/v1/post_info?code_or_id_or_url="
)

type SharingLink struct {
	Link    *string
	Comment *string
}

var rePatInsta *regexp.Regexp = regexp.MustCompile(PatInsta)
var reInsta *regexp.Regexp = regexp.MustCompile(videoPatInsta)
var code string

func ParseShort(link string, input tgbotapi.Update, ErrChan chan string) {
	DownloadYoutubeFile(link, input, ErrChan)
}

func ParseTikTok(link string, input tgbotapi.Update, ErrChan chan string) {
	DownloadTikTokFile(link, input, ErrChan)
}

func ParseReel(link string, input tgbotapi.Update, ErrChan chan string) {

	for _, element := range rePatInsta.FindAllStringSubmatch(link, -1) {
		code = element[1]
	}

	url := apiURL1 + code
	response, err := rapid.GetUrlReel(url, cfg.Config.RapidToken)

	if err != nil || response == "" {
		log.Printf("[ParseReel] Error while build url for rapidApi: %s\n", err)
		ErrChan <- fmt.Sprintf("[ParseReel] Error while build url for rapidApi: %s\n", err)
	}

	var capture string
	for _, element := range reInsta.FindAllStringSubmatch(response, -1) {
		capture = element[1]
	}
	DownloadReelFile(capture, input, ErrChan)
}

// func resolveRedirect(url string, ErrChan chan string) string {
// 	client := http.Client{
// 		CheckRedirect: func(req *http.Request, via []*http.Request) error {
// 			return http.ErrUseLastResponse
// 		},
// 	}
// 	resp, err := client.Get(url)
// 	if err != nil {
// 		ErrChan <- fmt.Sprintf("[resolveRedirect] Error in Get requst: %s\n", err)
// 		return ""
// 	}
// 	defer resp.Body.Close()

// 	location := resp.Header.Get("Location")
// 	if location == "" {
// 		return url
// 	}
// 	return location
// }

func DownloadTikTokFile(link string, input tgbotapi.Update, ErrChan chan string) {

	filename := fmt.Sprintf("%d.mp4", input.UpdateID)
	outputPath := filepath.Join("downloads", filename)

	cmd := exec.Command(
		"/app/bin/yt-dlp",
		"-f", "mp4",
		"-o", outputPath,
		link,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[DownloadTikTokFile] Error while downloading: %s\n%s\n", err, out)
		ErrChan <- fmt.Sprintf("[DownloadTikTokFile] Error while downloading: %s\n%s\n", err, out)
		return
	}
	log.Printf("[DownloadTikTokFile] Success: %s\n%s\n", filename, out)
}

func DownloadYoutubeFile(link string, input tgbotapi.Update, ErrChan chan string) {
	filename := fmt.Sprintf("%d.mp4", input.UpdateID)
	outputPath := filepath.Join("downloads", filename)

	cmd := exec.Command("/app/bin/yt-dlp", "-f", "mp4", "-o", outputPath, link)

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[DownloadYoutubeFile] Error while download youtube file: %s\n%s\n", err, out)
		ErrChan <- fmt.Sprintf("[DownloadYoutubeFile] Error while download youtube file: %s\n%s\n", err, out)
		return
	}
	log.Printf("[DownloadYoutubeFile] Successful downloaded file: %s\n%s\n", filename, out)
}

func DownloadReelFile(url string, input tgbotapi.Update, ErrChan chan string) {
	if url == "" {
		log.Printf("[DownloadReelFile] Error URL is empty field = %v\n", url)
		ErrChan <- fmt.Sprintf("[DownloadReelFile] Error URL is empty field = %v\n", url)
		return
	}

	res, err := http.Get(url)
	if err != nil {
		log.Printf("[DownloadReelFile] Error with GET request: %s\nStatus %s\n", err, res.Status)
		ErrChan <- fmt.Sprintf("[DownloadReelFile] Error with GET request: %s\nStatus %s\n", err, res.Status)
		return
	}
	defer res.Body.Close()

	filename := fmt.Sprintf("%d.mp4", input.UpdateID)
	path := filepath.Join("/app/downloads", filename)

	file, err := os.Create(path)
	if err != nil {
		log.Printf("[DownloadReelFile] Error while create download file: %s\n", err)
		ErrChan <- fmt.Sprintf("[DownloadReelFile] Error while create download file: %s\n", err)
		return
	}

	_, err = io.Copy(file, res.Body)
	if err != nil {
		log.Printf("[DownloadReelFile] Error while copy data in download file: %s\n", err)
		ErrChan <- fmt.Sprintf("[DownloadReelFile] Error while copy data in download file: %s\n", err)
		return
	}
}
