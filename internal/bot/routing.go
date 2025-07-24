// cSpell:disable
package bot

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kartmos/scraper-bot/internal/downloader"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	ffmpeg "github.com/u2takey/ffmpeg-go"
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
	case "help":
		helpText, err := loadText("./asserts/help.txt")
		if err != nil {
			log.Println("Ошибка чтения инструкции:", err)
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
	case "status":
		if input.Message.From.ID != adminID {
			msg.Text = "You not admin"
			bot.Send(msg)
			return
		}
		handleStatusCommand(input, bot)
	case "log":
		if input.Message.From.ID != adminID {
			msg.Text = "You not admin"
			bot.Send(msg)
			return
		}
		logfile := ("/app/logs/bot.log")
		doc := tgbotapi.NewDocument(adminID, tgbotapi.FilePath(logfile))
		bot.Send(doc)

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
	defer os.Remove(file)

	info, err := os.Stat(file)
	if err != nil {
		log.Printf("[videoSender] Error when bot try get size videofile: %s\n", err)
		ErrChan <- fmt.Sprintf("[videoSender] Error when bot try get size videofile: %s\n", err)
	}

	sizeMB := float64(info.Size()) / (1024 * 1024)
	finalFile := file
	if sizeMB > 50 {
		compressed := strings.TrimSuffix(file, ".mp4") + "_compressed.mp4"
		err := compressVideoFFMPEGGo(file, compressed)
		if err != nil {
			ErrChan <- fmt.Sprintf("[compressVideoFFMPEGGo] Ошибка сжатия: %s", err)
			return
		}
		defer os.Remove(compressed)
		finalFile = compressed
	}

	var videoMsg tgbotapi.VideoConfig
	videoMsg = BuildVideoMsg(input, finalFile, bridge)

	if _, err := bot.Send(videoMsg); err != nil {
		log.Printf("[videoSender]Error while bot try send video file %s\n", err)
		ErrChan <- fmt.Sprintf("[videoSender]Error while bot try send video file %s\n", err)
		SendErrorMessageChat(input, bot)
		return
	}
}

func compressVideoFFMPEGGo(inputPath, outputPath string) error {
	return ffmpeg.Input(inputPath).
		Output(outputPath, ffmpeg.KwArgs{
			"b:v":     "1M",
			"bufsize": "1M",
		}).
		OverWriteOutput().
		Run()
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

var startTime = time.Now()

func handleStatusCommand(input tgbotapi.Update, bot *tgbotapi.BotAPI) {
	uptime := time.Since(startTime).Truncate(time.Second)
	cpuPercents, _ := cpu.Percent(0, false)
	cpuUsage := 0.0
	if len(cpuPercents) > 0 {
		cpuUsage = cpuPercents[0]
	}

	memStats, _ := mem.VirtualMemory()
	diskStats, _ := disk.Usage("/")

	sessionCount := 0
	Queue.mu.Lock()
	for s := Queue.Head; s != nil; s = s.Next {
		sessionCount++
	}
	Queue.mu.Unlock()

	ip := "unknown"
	if resp, err := http.Get("https://api.ipify.org"); err == nil {
		defer resp.Body.Close()
		if body, err := io.ReadAll(resp.Body); err == nil {
			ip = string(body)
		}
	}

	msg := fmt.Sprintf(
		"📊 Bot Status\n\n"+
			"⏱ Uptime: %s\n"+
			"🧠 CPU: %.2f%%\n"+
			"💾 RAM: %.2f%% (%.2f GB total)\n"+
			"📀 Disk: %.2f%% (%.2f GB free)\n"+
			"👥 Active Sessions: %d\n"+
			"🌍 IP: %s",
		uptime,
		cpuUsage,
		memStats.UsedPercent,
		float64(memStats.Total)/1e9,
		diskStats.UsedPercent,
		float64(diskStats.Free)/1e9,
		sessionCount,
		ip,
	)

	bot.Send(tgbotapi.NewMessage(input.Message.Chat.ID, msg))
}
