// cSpell:disable
package bot

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	tgbotapi "github.com/kartmos/kartmos-telegram-bot-api/v5"
)

func ErrCollector(errChan chan string, bot *tgbotapi.BotAPI) {

	logFile, err := os.OpenFile("/app/logs/bot.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Не удалось открыть лог-файл: %s", err)
	}

	multi := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multi)

	msg := tgbotapi.NewMessage(adminID, "")
	timer := time.NewTimer(10 * time.Second)

	for {
		select {
		case err := <-errChan:
			msg.Text += fmt.Sprintf("\n[WARN]: %s\n", err)
			timer.Reset(10 * time.Second)

		case <-timer.C:
			if msg.Text != "" {
				bot.Send(msg)
				msg.Text = ""
			}

		}
	}
}

func SendErrorMessageChat(input tgbotapi.Update, bot *tgbotapi.BotAPI) {
	ChatID := input.Message.Chat.ID
	log.Printf("MSG in SendErrorMessageChat: %d\n", ChatID)
	msgErr := tgbotapi.NewMessage(ChatID, "")
	msgErr.MessageThreadID = input.Message.MessageThreadID
	msgErrAdmin := tgbotapi.NewMessage(adminID, "")
	msgErrAdmin.Text = fmt.Sprintf("Error in chat %v\n\nMessage: %s", input.FromChat().Title, input.Message.Text)
	msgErr.Text = "Пу-пу-пууу... Не получилось его скачать"
	bot.Send(msgErr)
	bot.Send(msgErrAdmin)
}
