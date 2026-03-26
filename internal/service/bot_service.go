package service

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

const maxTelegramVideoSize = 50 * 1024 * 1024

type Settings struct {
	AdminID     int64
	HelpFile    string
	WelcomeFile string
	CommandFile string
	LogFile     string
}

type Params struct {
	Chat     ChatGateway
	Videos   VideoRepository
	Settings Settings
	Started  time.Time
}

type BotService struct {
	chat      ChatGateway
	videos    VideoRepository
	adminID   int64
	helpFile  string
	welcome   string
	command   string
	logFile   string
	startedAt time.Time
	urlRegexp *regexp.Regexp
}

func NewBotService(p Params) *BotService {
	return &BotService{
		chat:      p.Chat,
		videos:    p.Videos,
		adminID:   p.Settings.AdminID,
		helpFile:  p.Settings.HelpFile,
		welcome:   p.Settings.WelcomeFile,
		command:   p.Settings.CommandFile,
		logFile:   p.Settings.LogFile,
		startedAt: p.Started,
		urlRegexp: regexp.MustCompile(`https?://\S+`),
	}
}

func (s *BotService) HandleMessage(ctx context.Context, msg IncomingMessage) error {
	s.handleWelcome(msg)

	if msg.IsCommand {
		return s.handleCommand(msg)
	}

	link := s.extractFirstURL(msg.Text)
	if link == "" {
		return nil
	}

	videoPath, err := s.videos.Download(ctx, msg.UpdateID, link)
	if err != nil {
		s.notifyDownloadError(msg, err)
		return err
	}
	defer func() {
		_ = os.Remove(videoPath)
	}()

	finalPath := videoPath
	compressedPath := strings.TrimSuffix(videoPath, ".mp4") + "_compressed.mp4"
	needCleanupCompressed := false

	if tooLarge(videoPath) {
		if err := compressVideo(videoPath, compressedPath); err == nil {
			finalPath = compressedPath
			needCleanupCompressed = true
		}
	}
	if needCleanupCompressed {
		defer func() {
			_ = os.Remove(compressedPath)
		}()
	}

	caption := fmt.Sprintf("Мем от %s", msg.UserDisplayName)
	if comment := extractComment(msg.Text); comment != "" {
		caption = caption + ":\n\n" + comment
	}

	if err := s.chat.SendVideo(msg.ChatID, msg.ThreadID, finalPath, caption); err != nil {
		s.notifyDownloadError(msg, err)
		return err
	}

	if msg.MessageID != 0 {
		_ = s.chat.DeleteMessage(msg.ChatID, msg.MessageID)
	}

	return nil
}

func (s *BotService) handleCommand(msg IncomingMessage) error {
	switch msg.Command {
	case "help":
		text, err := os.ReadFile(s.helpFile)
		if err != nil {
			return err
		}
		return s.chat.SendText(msg.ChatID, msg.ThreadID, string(text), true)
	case "start":
		text, err := os.ReadFile(s.welcome)
		if err != nil {
			return err
		}
		return s.chat.SendText(msg.ChatID, msg.ThreadID, string(text), true)
	case "status":
		if msg.UserID != s.adminID {
			return s.chat.SendText(msg.ChatID, msg.ThreadID, "You not admin", false)
		}
		return s.chat.SendText(msg.ChatID, msg.ThreadID, s.buildStatus(), false)
	case "log":
		if msg.UserID != s.adminID {
			return s.chat.SendText(msg.ChatID, msg.ThreadID, "You not admin", false)
		}
		return s.chat.SendDocument(s.adminID, msg.ThreadID, s.logFile)
	default:
		commands, err := os.ReadFile(s.command)
		if err != nil {
			return err
		}
		if err := s.chat.SendText(msg.ChatID, msg.ThreadID, "Такой команды не сущетсвует\n🗿Вот список моих команды:\n", false); err != nil {
			return err
		}
		return s.chat.SendText(msg.ChatID, msg.ThreadID, string(commands), true)
	}
}

func (s *BotService) handleWelcome(msg IncomingMessage) {
	if len(msg.NewChatMemberIDs) == 0 {
		return
	}

	for _, memberID := range msg.NewChatMemberIDs {
		if memberID != s.chat.BotID() {
			continue
		}
		text, err := os.ReadFile(s.welcome)
		if err != nil {
			return
		}
		_ = s.chat.SendText(msg.ChatID, msg.ThreadID, string(text), true)
		return
	}
}

func (s *BotService) notifyDownloadError(msg IncomingMessage, err error) {
	_ = s.chat.SendText(msg.ChatID, msg.ThreadID, "Пу-пу-пууу... Не получилось его скачать", false)
	_ = s.chat.SendText(s.adminID, 0, fmt.Sprintf("Error in chat %d\n\nMessage: %s\nError: %v", msg.ChatID, msg.Text, err), false)
}

func (s *BotService) buildStatus() string {
	uptime := time.Since(s.startedAt).Truncate(time.Second)

	cpuPercents, _ := cpu.Percent(0, false)
	cpuUsage := 0.0
	if len(cpuPercents) > 0 {
		cpuUsage = cpuPercents[0]
	}

	memStats, _ := mem.VirtualMemory()
	diskStats, _ := disk.Usage("/")

	return fmt.Sprintf(
		"📊 Bot Status\n\n"+
			"⏱ Uptime: %s\n"+
			"🧠 CPU: %.2f%%\n"+
			"💾 RAM: %.2f%% (%.2f GB total)\n"+
			"📀 Disk: %.2f%% (%.2f GB free)",
		uptime,
		cpuUsage,
		memStats.UsedPercent,
		float64(memStats.Total)/1e9,
		diskStats.UsedPercent,
		float64(diskStats.Free)/1e9,
	)
}

func (s *BotService) extractFirstURL(text string) string {
	match := s.urlRegexp.FindString(text)
	return strings.TrimSpace(match)
}

func extractComment(text string) string {
	if text == "" {
		return ""
	}
	parts := strings.SplitN(text, "\n", 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func tooLarge(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Size() > maxTelegramVideoSize
}

func compressVideo(inputPath, outputPath string) error {
	return ffmpeg.Input(inputPath).
		Output(outputPath, ffmpeg.KwArgs{
			"b:v":     "3.5M",
			"bufsize": "8M",
		}).
		OverWriteOutput().
		Run()
}
