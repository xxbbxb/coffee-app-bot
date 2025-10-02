package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/xxbbxb/coffee-app-bot/pkg/router"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	csOnlineReady    = "🟢Могу созвониться онлайн"
	csOnlineNotReady = "⚫Только оффлайн встречи"
)

func DialogEditName(bot *tgbotapi.BotAPI, chatId int64, back string) int {
	return SendMessage(bot, chatId, "Напишите ваше имя (по желанию прикрепите фото)",
		WithKeyboard(
			MenuVertical(tgbotapi.NewInlineKeyboardButtonData("Отмена", back)),
		),
	)
}

func DialogEditBio(bot *tgbotapi.BotAPI, chatId int64, back string) {
	SendMessage(bot, chatId, "Напишите о себе - работа, увлечения, навыки, интересы... (одним сообщением)",
		WithKeyboard(
			MenuVertical(tgbotapi.NewInlineKeyboardButtonData("Отмена", back)),
		),
	)
}

func DialogSettings(bot *tgbotapi.BotAPI, u *router.Update) int {
	settings := GetSettings(u.Context())
	var t []string
	var onlineButtonText string
	t = append(t, "Как и где вы предпочитаете пить кофе с сообщниками:\n")

	t = append(t, fmt.Sprintf("Видимость в поиске: <b>%d %%</b>", settings.Visibility))

	if settings.OnLineReady {
		t = append(t, fmt.Sprintf("Онлайн: <b>%s</b>", csOnlineReady))
		onlineButtonText = csOnlineNotReady
	} else {
		t = append(t, fmt.Sprintf("Онлайн: <b>%s</b>", csOnlineNotReady))
		onlineButtonText = csOnlineReady
	}

	if settings.OfflineCity == "" {
		settings.OfflineCity = "---"
	}
	t = append(t, fmt.Sprintf("Город(а) для оффлайн встреч: <b>%s</b>", settings.OfflineCity))

	return SendMessage(bot, u.ChatID(), strings.Join(t, "\n"),
		WithKeyboard(
			MenuVertical(
				tgbotapi.NewInlineKeyboardButtonData("📝Видимость", "/settings/visibility"),
				tgbotapi.NewInlineKeyboardButtonData(onlineButtonText, fmt.Sprintf("/settings/online/%t", !settings.OnLineReady)),
				tgbotapi.NewInlineKeyboardButtonData("🌍Указать город", "/settings/city/edit"),
				tgbotapi.NewInlineKeyboardButtonURL("💬Обратная связь", "https://forms.gle/SjKjrCKaPK1tMaHD9"),
				tgbotapi.NewInlineKeyboardButtonData("⬅️Назад", "/start"),
			),
		),
		WithReplaceMessage(u),
	)
}

func DialogVisibility(bot *tgbotapi.BotAPI, chatId int64, u *router.Update) int {
	t := []string{
		"0% - вы не будете видны в поиске",
		"1% - ваш профиль будет появляться в поиске в 100 раз реже чем другие профили",
		"50% - показывать профиль в 2 раза реже",
		"100% (значение по умолчанию) - изначально у всех одинаковые шансы",
		"200% - показывать профиль в 2 раза чаще",
	}

	return SendMessage(bot, chatId, strings.Join(t, "\n"),
		WithKeyboard(
			MenuHorizontal(
				tgbotapi.NewInlineKeyboardButtonData("0%", "/settings/visibility/0"),
				tgbotapi.NewInlineKeyboardButtonData("1%", "/settings/visibility/1"),
				tgbotapi.NewInlineKeyboardButtonData("50%", "/settings/visibility/50"),
				tgbotapi.NewInlineKeyboardButtonData("100%", "/settings/visibility/100"),
				tgbotapi.NewInlineKeyboardButtonData("200%", "/settings/visibility/200"),
				tgbotapi.NewInlineKeyboardButtonData("⬅️Назад", "/settings/show"),
			),
		),
		WithReplaceMessage(u),
	)
}

func DialogRandomProfile(bot *tgbotapi.BotAPI, ctx context.Context, chatId int64) int {
	user := GetRandomUser(ctx)
	if user != nil {
		kb := MenuVertical(
			tgbotapi.NewInlineKeyboardButtonData("☕Позвать на кофе", fmt.Sprintf("/match/invite/%d", user.Id)),
			tgbotapi.NewInlineKeyboardButtonData("🎲Пропустить", fmt.Sprintf("/match/deinvite/%d", user.Id)),
			tgbotapi.NewInlineKeyboardButtonData("❌Забанить", fmt.Sprintf("/match/reject/%d", user.Id)),
			tgbotapi.NewInlineKeyboardButtonData("⬅️В меню", "/start"),
		)
		return DialogProfile(bot, ctx, chatId, user, kb)
	}

	return SendMessage(bot, chatId, "🏁Вы видели всех сообщников и либо забанили. Приходите позднее когда появятся новые люди.",
		WithKeyboard(
			MenuVertical(
				tgbotapi.NewInlineKeyboardButtonData("⬅️В меню", "/start"),
			),
		),
	)
}

func DialogProfile(bot *tgbotapi.BotAPI, ctx context.Context, chatId int64, user *User, kb tgbotapi.InlineKeyboardMarkup, extraText ...string) int {

	if user == nil {
		dbUser, err := GetDB(ctx).GetUser(chatId)
		if err != nil {
			log.WithError(err).Error("show profile dialog error")
		}
		user = &dbUser
	}
	if user.PhotoFileId != "" {
		return SendMessage(bot, chatId, RenderProfileText(ctx, *user, extraText...),
			WithKeyboard(kb),
			WithPhoto(user.PhotoFileId),
		)
	}
	return SendMessage(bot, chatId, RenderProfileText(ctx, *user, extraText...),
		WithKeyboard(kb),
	)
}

func RenderProfileText(ctx context.Context, user User, extraText ...string) string {
	text := []string{}

	text = append(text, fmt.Sprintf("<b>%s</b>", user.ShownName))
	if user.Bio != "" {
		text = append(text, user.Bio)
	}

	var availability string
	settings, err := GetDB(ctx).GetSettings(user.Id)
	if err != nil {
		log.WithError(err).WithField("user-id", user.Id).Warn("unable to read user settings")
	}
	if settings != nil {
		if settings.OnLineReady {
			availability = csOnlineReady
		}
		if availability != "" && settings.OfflineCity != "" {
			availability = fmt.Sprintf("%s или ", availability)
		}
		if settings.OfflineCity != "" {
			availability = fmt.Sprintf("%sоффлайн встречи: %s", availability, settings.OfflineCity)
		}
		text = append(text, fmt.Sprintf("\n---\n<i>%s</i>", availability))
	}

	return strings.Join(text, "\n")
}
