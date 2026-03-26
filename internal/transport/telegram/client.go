package telegram

import (
	"context"

	tgbotapi "github.com/kartmos/kartmos-telegram-bot-api/v5"
)

type Client struct {
	api *tgbotapi.BotAPI
}

func NewClient(token string) (*Client, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &Client{api: api}, nil
}

func (c *Client) Run(ctx context.Context, handler UpdateHandler) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := c.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return nil
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			if err := handler.HandleUpdate(ctx, update); err != nil {
				continue
			}
		}
	}
}

func (c *Client) BotID() int64 {
	return c.api.Self.ID
}

func (c *Client) SendText(chatID int64, threadID int, text string, markdown bool) error {
	msg := tgbotapi.NewMessage(chatID, text)
	if threadID != 0 {
		msg.MessageThreadID = threadID
	}
	if markdown {
		msg.ParseMode = "Markdown"
	}
	_, err := c.api.Send(msg)
	return err
}

func (c *Client) SendVideo(chatID int64, threadID int, filePath string, caption string) error {
	video := tgbotapi.NewVideo(chatID, tgbotapi.FilePath(filePath))
	video.Caption = caption
	if threadID != 0 {
		video.MessageThreadID = threadID
	}
	_, err := c.api.Send(video)
	return err
}

func (c *Client) SendDocument(chatID int64, threadID int, filePath string) error {
	doc := tgbotapi.NewDocument(chatID, tgbotapi.FilePath(filePath))
	if threadID != 0 {
		doc.MessageThreadID = threadID
	}
	_, err := c.api.Send(doc)
	return err
}

func (c *Client) DeleteMessage(chatID int64, messageID int) error {
	if messageID == 0 {
		return nil
	}
	_, err := c.api.Request(tgbotapi.DeleteMessageConfig{ChatID: chatID, MessageID: messageID})
	return err
}
