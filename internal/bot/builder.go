package bot

import (
	"fmt"
	"strings"

	tgbotapi "github.com/kartmos/kartmos-telegram-bot-api/v5"
)

func BuildDelConfig(input []tgbotapi.Update, delbridge chan<- []tgbotapi.DeleteMessageConfig) {
	deleteConfig := make([]tgbotapi.DeleteMessageConfig, len(input))
	for i, update := range input {
		deleteConfig[i] = tgbotapi.DeleteMessageConfig{
			ChatID:    update.Message.Chat.ID,
			MessageID: update.Message.MessageID,
		}
	}
	delbridge <- deleteConfig
	close(delbridge)
}

func BuildVideoMsg(input tgbotapi.Update, path string, bridge <-chan string) tgbotapi.VideoConfig {
	var msg tgbotapi.VideoConfig
	var comment string
	link := ExtcractDomain(input.Message.Text)
	switch link {
	case "vt.tiktok.com":
		comment = ParseCommentTikTok(input)
		msg = buildTikTokMsg(input, path, comment)
	case "youtube.com":
		comment = ParseCommentShorts(input)
		msg = buildShortMsg(input, path, comment)
	case "www.instagram.com":
		comment = ParseCommentReel(bridge)
		msg = buildReelMsg(input, path, comment)
	}

	return msg 
}

func ParseCommentReel(bridge <-chan string) string {
	select {
	case val, ok := <-bridge:
		if !ok {
			return ""
		}
		return val
	default:
		return ""
	}
}

func ParseCommentShorts(input tgbotapi.Update) string {
	parts := strings.SplitN(input.Message.Text, "\n", 2)
	var comment string
	if len(parts) == 2 {
		comment = parts[1]
		fmt.Println(comment)
	} else {
		return ""
	}
	return comment
}

func ParseCommentTikTok(input tgbotapi.Update) string {
	
	parts := strings.SplitN(input.Message.Text, "\n\n", 2)
	var comment string
	if len(parts) == 2 {
		comment = parts[1]
	} else {
		return ""
	}
	return comment
}

func buildReelMsg(input tgbotapi.Update, path string, comment string) tgbotapi.VideoConfig {

	videoMsg := tgbotapi.NewVideo(input.Message.Chat.ID, tgbotapi.FilePath(path))
	videoMsg.Caption = fmt.Sprintf("Мем от %s:\n\n%s", input.Message.From, comment)
	return videoMsg
}

func buildShortMsg(input tgbotapi.Update, path string, comment string) tgbotapi.VideoConfig {

	videoMsg := tgbotapi.NewVideo(input.Message.Chat.ID, tgbotapi.FilePath(path))
	videoMsg.Caption = fmt.Sprintf("Мем от %s:\n\n%s", input.Message.From, comment)
	return videoMsg
}
func buildTikTokMsg(input tgbotapi.Update, path string, comment string) tgbotapi.VideoConfig {

	videoMsg := tgbotapi.NewVideo(input.Message.Chat.ID, tgbotapi.FilePath(path))
	videoMsg.Caption = fmt.Sprintf("Мем от %s:\n\n%s", input.Message.From, comment)
	return videoMsg
}
