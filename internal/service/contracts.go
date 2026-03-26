package service

import "context"

type IncomingMessage struct {
	UpdateID         int
	MessageID        int
	ChatID           int64
	ThreadID         int
	UserID           int64
	UserDisplayName  string
	Text             string
	IsCommand        bool
	Command          string
	NewChatMemberIDs []int64
}

type Handler interface {
	HandleMessage(ctx context.Context, msg IncomingMessage) error
}

type ChatGateway interface {
	BotID() int64
	SendText(chatID int64, threadID int, text string, markdown bool) error
	SendVideo(chatID int64, threadID int, filePath string, caption string) error
	SendDocument(chatID int64, threadID int, filePath string) error
	DeleteMessage(chatID int64, messageID int) error
}

type VideoRepository interface {
	Download(ctx context.Context, updateID int, sourceLink string) (string, error)
}
