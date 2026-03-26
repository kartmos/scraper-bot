package telegram

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/kartmos/kartmos-telegram-bot-api/v5"
	"github.com/kartmos/scraper-bot/internal/service"
)

type UpdateHandler interface {
	HandleUpdate(ctx context.Context, update tgbotapi.Update) error
}

type Handler struct {
	service service.Handler
}

func NewHandler(botService service.Handler) *Handler {
	return &Handler{service: botService}
}

func (h *Handler) HandleUpdate(ctx context.Context, update tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}

	return h.service.HandleMessage(ctx, mapMessage(update))
}

func mapMessage(update tgbotapi.Update) service.IncomingMessage {
	msg := service.IncomingMessage{
		UpdateID:        update.UpdateID,
		MessageID:       update.Message.MessageID,
		ChatID:          update.Message.Chat.ID,
		ThreadID:        update.Message.MessageThreadID,
		UserID:          update.Message.From.ID,
		UserDisplayName: formatDisplayName(update.Message.From),
		Text:            update.Message.Text,
		IsCommand:       update.Message.IsCommand(),
		Command:         update.Message.Command(),
	}

	if update.Message.NewChatMembers != nil {
		members := make([]int64, 0, len(update.Message.NewChatMembers))
		for _, member := range update.Message.NewChatMembers {
			members = append(members, member.ID)
		}
		msg.NewChatMemberIDs = members
	}

	return msg
}

func formatDisplayName(user *tgbotapi.User) string {
	if user == nil {
		return "unknown"
	}
	if user.UserName != "" {
		return "@" + user.UserName
	}

	fullName := strings.TrimSpace(strings.Join([]string{user.FirstName, user.LastName}, " "))
	if fullName != "" {
		return fullName
	}

	return fmt.Sprintf("user_%d", user.ID)
}
