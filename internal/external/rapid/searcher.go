package rapid

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	host = "instagram-scraper-api2.p.rapidapi.com"
)

var client = &http.Client{
	Timeout: 5 * time.Second,
}

func GetUrlReel(url string, rapidTokens []string) (string, error) {

	var response string
	for idx, token := range rapidTokens {

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("[GetUrlReel] Token: %d\nError while create new request: %s\n", idx, err)
			continue
		}

		// log.Printf("\n====== [GetUrlReel] REQUEST ======\n\n%v\n\n==================\n", req)

		req.Header.Set("x-rapidapi-key", token)
		req.Header.Set("x-rapidapi-host", host)

		res, err := client.Do(req)
		if err != nil {
			log.Printf("[GetUrlReel] Token: %d\nResponse status: %s\nError while send request: %s\n", idx, res.Status, err)
			continue
		}

		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Printf("[GetUrlReel] Token: %d\nError while read response's body: %s\n", idx, err)
			continue
		}

		response = string(body)
		// log.Printf("\n====== [GetUrlReel] RESPONSE ======\n\n%v\n\n==================\n", response)
		return response, nil

	}
	err := fmt.Errorf("[GetUrlReel] All tokens failed or no valid response")
	return "", err
}
