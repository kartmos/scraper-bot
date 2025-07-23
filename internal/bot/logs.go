// cSpell:disable
package bot

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
			msg.Text += fmt.Sprintf("[WARN]: %s\n", err)
			timer.Reset(10 * time.Second)

		case <-timer.C:
			if msg.Text != "" {
				bot.Send(msg)
				SendErrorMessageChat(bot)
				msg.Text = ""

			}

		}
	}
}

func SendErrorMessageChat(bot *tgbotapi.BotAPI) {
	ChatID := <-ChatIdChan
	log.Printf("MSG in SendErrorMessageChat: %d", ChatID)
	msgChat := tgbotapi.NewMessage(ChatID, "")
	msgChat.Text = "Что-то пошло не так. Пожалуйста, попробуйте позже."
	bot.Send(msgChat)
}
