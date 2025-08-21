// cSpell:disable
package bot

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/kartmos/kartmos-telegram-bot-api/v5"
	cfg "github.com/kartmos/scraper-bot/config"
)

var Queue = NewQueue()
var ErrChan chan string

const (
	sessionLive = 2 * time.Second
)

type UserSession struct {
	UserID int64
	Val    []tgbotapi.Update
	mu     sync.Mutex
	Timer  *time.Timer
	Prev   *UserSession
	Next   *UserSession
}

type QueueSessionList struct {
	mu   sync.Mutex
	Head *UserSession
}

func NewQueue() *QueueSessionList {
	return &QueueSessionList{}
}

func StartBot() {

	ErrChan = make(chan string)
	defer close(ErrChan)

	cfg.Config.CheckConfig()
	log.Println("[DONE] Set options from config file")
	token := os.Getenv("BOT_TOKEN")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Printf("[Start] Bot doesn't born: %s", err)
		ErrChan <- fmt.Sprintf("[Start] Bot doesn't born: %s", err)
	}

	bot.Debug = true

	go ErrCollector(ErrChan, bot)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
 
	for update := range updates {

		if update.Message == nil {
			continue
		}

		log.Printf("[Get Update] ChatID: %d | UserID: %d | Text: %q",
			update.Message.Chat.ID,
			update.Message.From.ID,
			update.Message.Text)

		find := finderSession(update)
		if !find {
			Queue.appendNewSession(update, bot)
			log.Printf("\n\n[NewSession] Создана сессия для %d\n\n", update.Message.From.ID)
		}
		if update.Message != nil && update.Message.NewChatMembers != nil {
			for _, user := range update.Message.NewChatMembers {
				if user.ID == bot.Self.ID {
					welcome, err := loadText("./asserts/welcome.txt")
					if err != nil {
						log.Println("Ошибка чтения приветствия:", err)
						continue
					}

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, welcome)
					msg.MessageThreadID=update.Message.MessageThreadID
					msg.ParseMode = "Markdown"
					bot.Send(msg)
				}
			}

		}
	}
}

func loadText(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения welcome файла: %w", err)
	}
	return string(content), nil
}

func NewSession(update tgbotapi.Update, bot *tgbotapi.BotAPI) *UserSession {
	val := []tgbotapi.Update{update}
	session := &UserSession{
		UserID: update.Message.From.ID,
		Val:    val,
		Next:   nil,
	}
	session.Timer = time.AfterFunc(sessionLive, func() {
		processing(session.Val, bot)
		deleteFromQueue(session)
	})
	return session
}

// delete session after deadline
func deleteFromQueue(session *UserSession) {

	if session == nil {
		return
	}
	
	Queue.mu.Lock()
	defer Queue.mu.Unlock()
	if Queue.Head == nil {
		return
	}
	curr := Queue.Head
	for curr != nil {
		if curr == session {
			if curr.Prev != nil {
				curr.Prev.Next = curr.Next
			} else {
				Queue.Head = curr.Next
			}
			if curr.Next != nil {
				curr.Next.Prev = curr.Prev
			}
			return
		}
		curr = curr.Next
	}
}

// creat new session at the end of QueueSessionList
func (q *QueueSessionList) appendNewSession(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	newSession := NewSession(update, bot)

	q.mu.Lock()
	defer q.mu.Unlock()

	if q.Head == nil {
		q.Head = newSession
		return
	}

	curr := q.Head
	for curr.Next != nil {
		curr = curr.Next
	}

	newSession.Prev = curr
	curr.Next = newSession
}

func finderSession(update tgbotapi.Update) bool {
	if Queue.Head == nil {
		return false
	}
	curr := Queue.Head

	for curr != nil {
		if curr.UserID == update.Message.From.ID {
			curr.mu.Lock()
			curr.Val = append(curr.Val, update)
			log.Printf("\n\n[finderSession] Добавлен второй update для %d. len=%d\n\n", curr.UserID, len(curr.Val))
			curr.mu.Unlock()
			return true
		}
		curr = curr.Next
	}
	return false
}

func processing(input []tgbotapi.Update, bot *tgbotapi.BotAPI) {
	var idx int
	bridge := make(chan string, 1)
	delbridge := make(chan []tgbotapi.DeleteMessageConfig, 1)

	if len(input) == 2 {
		idx = ScannerUpdate(input, bridge)
	} else {
		idx = 0
	}
	BuildDelConfig(input, delbridge)
	RouteMessege(input[idx], bot, bridge, delbridge)

	// if input.CallbackQuery != nil {
	// 	HandleCallback(input, bot)
	// }
}

func ScannerUpdate(input []tgbotapi.Update, bridge chan string) int {
	buffer1 := input[0]
	buffer2 := input[1]
	defer close(bridge)

	if strings.Contains(buffer1.Message.Text, FindStr) {
		bridge <- buffer2.Message.Text
		return 0
	} else {
		bridge <- buffer1.Message.Text
		return 1
	}
}

// func conditionCheckReels(input tgbotapi.Update, buffer tgbotapi.Update) bool {
// 	if buffer.Message.Date != input.Message.Date {
// 		log.Printf("\n\n 1 \n\n\n")
// 		return false
// 	}
// 	if buffer.Message.Chat.ID != input.Message.Chat.ID {
// 		log.Printf("\n\n 2 \n\n\n")
// 		return false
// 	}
// 	if int64(buffer.Message.From.ID) != int64(input.Message.From.ID) {
// 		log.Printf("\n\n 3 \n\n\n")
// 		return false
// 	}
// 	log.Printf("\n\n Пошло дальше \n\n\n")
// 	return true
// }
