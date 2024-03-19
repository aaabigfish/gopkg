package tgbotapi

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/aaabigfish/gopkg/config"
	"github.com/aaabigfish/gopkg/log"
)

type BotApi tgbotapi.BotAPI

type Bot struct {
	botapi  *tgbotapi.BotAPI
	chatIds []int64
}

func NewBot(token string, chatIds ...int64) *Bot {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Errorf("tgbotapi.NewBotAPI(%s) err(%v)", token, err)
		return &Bot{}
	}

	if config.Env == config.EnvDev {
		bot.Debug = true
	}

	return &Bot{botapi: bot, chatIds: chatIds}
}

func (b *Bot) GetBotApi() *tgbotapi.BotAPI {
	return b.botapi
}

func (b *Bot) SendMsg(message string) {
	if b.botapi == nil {
		return
	}
	for _, chatId := range b.chatIds {
		msg := tgbotapi.NewMessage(chatId, message)

		if m, err := b.botapi.Send(msg); err != nil {
			log.Errorf("tgbotapi SendMsg(%s), msg(%+v)  err(%v)", message, m, err)
		}
	}

	return
}

func (b *Bot) SendMarkdownMsg(message string) {
	if b.botapi == nil {
		return
	}
	for _, chatId := range b.chatIds {
		msg := tgbotapi.NewMessage(chatId, message)
		msg.ParseMode = tgbotapi.ModeMarkdown

		if m, err := b.botapi.Send(msg); err != nil {
			log.Errorf("tgbotapi SendMsg(%s), msg(%+v)  err(%v)", message, m, err)
		}
	}

	return
}

func (b *Bot) Send(chatId int64, message string) error {
	if b.botapi == nil {
		return nil
	}
	msg := tgbotapi.NewMessage(chatId, message)

	_, err := b.botapi.Send(msg)
	return err
}

func (b *Bot) SendChannelMsg(chanName string, message string) error {
	if b.botapi == nil {
		return nil
	}
	msg := tgbotapi.NewMessageToChannel(chanName, message)

	_, err := b.botapi.Send(msg)
	return err
}
