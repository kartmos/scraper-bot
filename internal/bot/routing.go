// cSpell:disable
package bot

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kartmos/bot-insta/internal/downloader"
)

const (
	FindStr         = "https"
	extDomain       = `https://([^/]+)`
	adminID   int64 = 291182090
	videoPath       = "./downloads"
)

func ExtcractDomain(url string) string {
	re := regexp.MustCompile(extDomain)
	match := re.FindStringSubmatch(url)
	return match[1]
}

func RouteMessege(input tgbotapi.Update, bot *tgbotapi.BotAPI, bridge <-chan string, delbridge chan []tgbotapi.DeleteMessageConfig) {

	if input.Message.IsCommand() {
		commandHendler(input, bot)
		return
	}

	if strings.Contains(input.Message.Text, FindStr) {
		link := ExtcractDomain(input.Message.Text)
		switch link {
		case "vt.tiktok.com":
			downloader.ParseTikTok(input.Message.Text, input, ErrChan)
			Cleaner(delbridge, bot)
		case "youtube.com":
			downloader.ParseShort(input.Message.Text, input, ErrChan)
			Cleaner(delbridge, bot)
		case "www.instagram.com":
			downloader.ParseReel(input.Message.Text, input, ErrChan)
			Cleaner(delbridge, bot)
		default:
			return
		}
		videoSender(input, bot, bridge)
	}

}

func Cleaner(delbridge chan []tgbotapi.DeleteMessageConfig, bot *tgbotapi.BotAPI) {
	deleteConfig := <-delbridge
	log.Printf("[Cleaner] Get from chan: %v", deleteConfig)
	for _, val := range deleteConfig {
		if val.MessageID != 0 {
			if _, err := bot.Request(val); err != nil {
				log.Printf("[cleaner] Error deleting link message: %s", err)
			}
		}
	}
}

func commandHendler(input tgbotapi.Update, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(input.Message.Chat.ID, "")
	switch input.Message.Command() {
	// case "adminID":
	// 	if adminID == 0 && input.Message.Text == "/admin 9f7aD!4tKz" {
	// 		adminID = input.Message.From.ID
	// 		msg.Text = "Теперь ты стал админом"
	// 		bot.Send(msg)
	// 	}
	// 	if input.Message.From.ID != adminID {
	// 		msg.Text = "Пу-пу-пу. А админ уже назначен."
	// 		bot.Send(msg)
	// 	}
	case "help":
		helpText, err := loadText("./asserts/help.txt")
		if err != nil {
			log.Println("Ошибка чтения приветствия:", err)
			ErrChan <- fmt.Sprintf("[WARN]Error when bot try read file help.txt: %s\n", err)
		}

		msg := tgbotapi.NewMessage(input.Message.Chat.ID, helpText)
		msg.ParseMode = "Markdown"
		bot.Send(msg)

		// photo := tgbotapi.NewPhoto(input.Message.Chat.ID, tgbotapi.FilePath("./asserts/help.jpg"))
		// bot.Send(photo)
	case "start":
		welcome, err := loadText("./asserts/welcome.txt")
		if err != nil {
			log.Println("Ошибка чтения приветствия:", err)
		}

		msg := tgbotapi.NewMessage(input.Message.Chat.ID, welcome)
		msg.ParseMode = "Markdown"
		bot.Send(msg)
	default:
		msg.Text = "Такой команды не сущетсвует\nВот список моих команды:\n/help..."
		bot.Send(msg)
		//будет дописываться
	}
}

func HandleCallback(input tgbotapi.Update, bot *tgbotapi.BotAPI) {

	if input.CallbackQuery.Data == "poke_dev" {
		logMsg := fmt.Sprintf("⚠️ Пользователь @%s нажал кнопку 'Пнуть разработчика'", input.CallbackQuery.From.UserName)
		alert := tgbotapi.NewMessage(adminID, logMsg)
		bot.Send(alert)

		notify := tgbotapi.NewCallback(input.CallbackQuery.ID, "Сообщение разработчику отправлено!")
		bot.Request(notify)
	}
}

func videoSender(input tgbotapi.Update, bot *tgbotapi.BotAPI, bridge <-chan string) {

	file := videoFileFinder(videoPath, input)
	if file != "" {
		defer os.Remove(file)
	}

	var videoMsg tgbotapi.VideoConfig
	videoMsg = BuildVideoMsg(input, file, bridge)

	if _, err := bot.Send(videoMsg); err != nil {
		log.Printf("[videoSender]Error while bot try send video file %s\n", err)
		ErrChan <- fmt.Sprintf("[videoSender]Error while bot try send video file %s\n", err)
		return
	}
}

func videoFileFinder(dir string, update tgbotapi.Update) string {
	updateID := strconv.Itoa(update.UpdateID)
	targetName := updateID + ".mp4"
	fullPath := filepath.Join(dir, targetName)

	if _, err := os.Stat(fullPath); err == nil {
		return fullPath
	} else {
		log.Printf("[videoFileFinder] File %s not found: %v\n", fullPath, err)
		ErrChan <- fmt.Sprintf("[videoFileFinder] File %s not found: %v\n", fullPath, err)
		return ""
	}
}

// func RouteParserMethod(input tgbotapi.Update, link string, resMsg tgbotapi.MessageConfig, bot *tgbotapi.BotAPI) {
// 	videoPath, err := d.DownloadFile(link)
// 	if err != nil {
// 		resMsg.Text = "Я сломався"
// 	}
// 	defer os.Remove(videoPath)
// 	video := tgbotapi.NewVideo(input.Message.Chat.ID, tgbotapi.FilePath(videoPath))
// 	resMsg.Text = "!!!ВНИМАНИЕ МЕМ!!!"
// 	if _, err := bot.Send(msg); err != nil {
// 		log.Panic(err)
// 	}
// 	if _, err := bot.Send(video); err != nil {
// 		resMsg.Text = "Видно был не очень рилс и я его не донес"
// 	}
// }
