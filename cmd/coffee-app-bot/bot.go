package main

import (
	"coffee-app-bot/pkg/router"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func MenuVertical(buttons ...tgbotapi.InlineKeyboardButton) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, b := range buttons {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{b})
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func MenuHorizontal(buttons ...tgbotapi.InlineKeyboardButton) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttons...))
}

type MessageToSend struct {
	photoFileId string
	kb          *tgbotapi.InlineKeyboardMarkup
	msgToUpdate int
}
type SendMessageOption func(*MessageToSend)

func WithKeyboard(kb tgbotapi.InlineKeyboardMarkup) func(*MessageToSend) {
	return func(m *MessageToSend) {
		m.kb = &kb
	}
}

func WithPhoto(photoFileId string) func(*MessageToSend) {
	return func(m *MessageToSend) {
		m.photoFileId = photoFileId
	}
}

func WithReplaceMessage(u *router.Update) func(*MessageToSend) {
	return func(m *MessageToSend) {
		if u.CallbackQuery != nil && len(u.CallbackQuery.Message.Photo) == 0 {
			m.msgToUpdate = u.CallbackQuery.Message.MessageID
		}

	}
}

func SendMessage(bot *tgbotapi.BotAPI, chatId int64, text string, opts ...SendMessageOption) int {

	m := MessageToSend{}
	for _, o := range opts {
		o(&m)
	}

	if m.photoFileId != "" {
		photo := tgbotapi.NewPhoto(chatId, tgbotapi.FileID(m.photoFileId))
		photo.Caption = text
		photo.ParseMode = tgbotapi.ModeHTML
		if m.kb != nil {
			photo.ReplyMarkup = m.kb
		}

		return doSend(bot, photo)
	}
	if m.msgToUpdate != 0 {
		var edit tgbotapi.EditMessageTextConfig
		if m.kb != nil {
			edit = tgbotapi.NewEditMessageTextAndMarkup(chatId, m.msgToUpdate, text, *m.kb)
		} else {
			edit = tgbotapi.NewEditMessageText(chatId, m.msgToUpdate, text)
		}
		edit.ParseMode = tgbotapi.ModeHTML
		return doSend(bot, edit)
	}

	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = tgbotapi.ModeHTML
	if m.kb != nil {
		msg.ReplyMarkup = m.kb
	}
	return doSend(bot, msg)
}

func doSend(bot *tgbotapi.BotAPI, c tgbotapi.Chattable) int {
	resp, err := bot.Send(c)
	if err != nil {
		log.WithError(err).Error("unable to send message")
		return 0
	}
	return resp.MessageID
}

func RejectIncomingMessage(bot *tgbotapi.BotAPI, u *router.Update, reason string) {
	log.WithField("user-id", u.ChatID()).WithField("route", u.RoutePath()).Debugf("New rejection: %s", reason)
	msg := tgbotapi.NewMessage(u.ChatID(), fmt.Sprintf("❗%s", reason))
	msg.ReplyToMessageID = u.Message.MessageID
	_, err := bot.Send(msg)
	if err != nil {
		log.WithField("user-id", u.ChatID()).WithField("route", u.RoutePath()).WithError(err).Error("unable to send rejection message")
	}
}
