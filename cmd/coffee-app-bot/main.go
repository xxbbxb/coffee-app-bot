package main

import (
	"fmt"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"

	"github.com/xxbbxb/coffee-app-bot/pkg/router"
)

var log = logrus.New()

func init() {
	log.SetFormatter(&logrus.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)

}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("COFFEE_TELEGRAM_APITOKEN"))
	if err != nil {
		panic(err)
	}
	if os.Getenv("COFFEE_DEBUG") != "" {
		bot.Debug = true
	}
	updateConfig := tgbotapi.NewUpdate(0)

	updateConfig.Timeout = 30

	r := router.NewRouter(func(u *router.Update) {
		chatId := u.ChatID()
		msg := tgbotapi.NewMessage(chatId, "Я не знаю что делать с этой информацией нажмите /start чтобы открыть меню")
		if u.Message != nil {
			msg.ReplyToMessageID = u.Message.MessageID
			bot.Send(msg)
		}
	})

	r.Use(CoffeeDBMiddleware(log, os.Getenv("COFFEE_DB_DSN")))
	r.Use(router.AutoCallbackOK(bot))
	r.CallbackDataHandler("/cancel", func(u *router.Update) {
		if u.CallbackQuery != nil {
			del := tgbotapi.NewDeleteMessage(u.ChatID(), u.CallbackQuery.Message.MessageID)
			bot.Send(del)
		}
	})
	r.Route("/match", func(r router.Router) {
		r.CallbackDataHandler("/search", func(u *router.Update) {
			DialogRandomProfile(bot, u.Context(), u.ChatID())
		})
		r.CallbackDataHandler("/invite/{value}", func(u *router.Update) {
			user := GetUser(u.Context())
			contactIdStr := router.GetParam(u.Context(), "value")
			contactId, _ := strconv.ParseInt(contactIdStr, 10, 64)
			contact, err := DoCoffeeUser(u.Context(), contactId)
			if err != nil {
				log.WithError(err).WithField("user-id", u.ChatID()).WithField("contact-id", contactId).Error("Unable to DoCoffeeUser")
			}
			if contact != nil {
				var msg tgbotapi.MessageConfig
				msg = tgbotapi.NewMessage(u.ChatID(), fmt.Sprintf("Новое совпадение🔥\nНапишите пользователю %s (@%s) и договоритесь когда сможете попить кофе. Приятной встречи🤘", contact.ShownName, contact.Login))
				bot.Send(msg)
				msg = tgbotapi.NewMessage(contact.Id, fmt.Sprintf("Новое совпадение🔥\nНапишите пользователю %s (@%s) и договоритесь когда сможете попить кофе. Приятной встречи🤘", user.ShownName, user.Login))
				bot.Send(msg)
				return
			}
			DialogProfile(bot, u.Context(), contactId, &user, MenuVertical(
				tgbotapi.NewInlineKeyboardButtonData("☕Принять приглашение", fmt.Sprintf("/match/invite/%d", user.Id)),
				tgbotapi.NewInlineKeyboardButtonData("🗑️Удалить приглашение", fmt.Sprintf("/match/deinvite/%d", user.Id)),
				tgbotapi.NewInlineKeyboardButtonData("❌Забанить", fmt.Sprintf("/match/reject/%d", user.Id)),
				tgbotapi.NewInlineKeyboardButtonData("⬅️В меню", "/start"),
			), "🌠Новое приглашение, воспользуйтесь кнопками чтобы ответить")
			DialogRandomProfile(bot, u.Context(), u.ChatID())
		})
		r.CallbackDataHandler("/reject/{value}", func(u *router.Update) {
			contactIdStr := router.GetParam(u.Context(), "value")
			contactId, _ := strconv.ParseInt(contactIdStr, 10, 64)
			RejectUser(u.Context(), contactId)
			del := tgbotapi.NewDeleteMessage(u.ChatID(), u.CallbackQuery.Message.MessageID)
			bot.Send(del)
			DialogRandomProfile(bot, u.Context(), u.ChatID())
		})
		r.CallbackDataHandler("/deinvite/{value}", func(u *router.Update) {
			contactIdStr := router.GetParam(u.Context(), "value")
			contactId, _ := strconv.ParseInt(contactIdStr, 10, 64)
			DeinviteUser(u.Context(), contactId)
			del := tgbotapi.NewDeleteMessage(u.ChatID(), u.CallbackQuery.Message.MessageID)
			bot.Send(del)
			DialogRandomProfile(bot, u.Context(), u.ChatID())
		})
	})

	r.Route("/profile", func(r router.Router) {
		kbProfile := MenuVertical(
			tgbotapi.NewInlineKeyboardButtonData("📝Изменить имя и фото", "/profile/name/edit"),
			tgbotapi.NewInlineKeyboardButtonData("📝Изменить информацию о себе", "/profile/bio/edit"),
			tgbotapi.NewInlineKeyboardButtonData("🗑️Удалить профиль", "/profile/delete/confirm"),
			tgbotapi.NewInlineKeyboardButtonData("⬅️В меню", "/start"),
		)
		r.CallbackDataHandler("/show", func(u *router.Update) {
			user := GetUser(u.Context())
			DialogProfile(bot, u.Context(), u.ChatID(), &user, kbProfile)
		})
		r.CallbackDataHandler("/delete/confirm", func(u *router.Update) {
			SendMessage(bot, u.ChatID(), "Да удалить все мои данные: профиль, контакты и настройки",
				WithKeyboard(
					MenuVertical(
						tgbotapi.NewInlineKeyboardButtonData("❌Да, удалить", u.NeighborRoute("/delete/destroy")),
						tgbotapi.NewInlineKeyboardButtonData("⬅️Отмена", u.NeighborRoute("/show")),
					),
				),
			)
		})
		r.CallbackDataHandler("/delete/destroy", func(u *router.Update) {
			DeleteUser(u.Context(), GetUser(u.Context()))
			SendMessage(bot, u.ChatID(), "Спасибо за то что были с нами")
		})
		r.CallbackDataHandler("/name/edit", func(u *router.Update) {
			DialogEditName(bot, u.ChatID(), "/cancel")
			r.Expect(u, u.NeighborRoute("/name/input"))
		})
		r.StateHandler("/name/input", func(u *router.Update) {
			ctx := u.Context()
			if u.Message == nil {
				r.Expect(u, u.NeighborRoute("/name/input"))
				RejectIncomingMessage(bot, u, "Попробуйте снова: сообщение не содержит имени и фотографии")
				return
			}

			user := GetUser(ctx)
			if len(u.Message.Photo) > 0 {
				user.PhotoFileId = u.Message.Photo[len(u.Message.Photo)-1].FileID
			}
			if u.Message.Caption != "" {
				user.ShownName = u.Message.Caption
			}
			if u.Message.Text != "" {
				user.ShownName = u.Message.Text
			}

			if user.ShownName == "" {
				r.Expect(u, u.NeighborRoute("/name/input"))
				RejectIncomingMessage(bot, u, "Попробуйте снова: вы не указали своё имя")
				return
			}

			if err := AddOrUpdateUser(ctx, user); err != nil {
				log.WithError(err).WithField("user-id", u.ChatID()).Error("Unable to save users shownName")
				RejectIncomingMessage(bot, u, "Попробуйте позже: произошла ошибка")
				return
			}

			if !IsUserProfileCompleted(user) {
				r.Expect(u, u.NeighborRoute("/bio/input"))
				DialogEditBio(bot, u.ChatID(), "/cancel")
			} else {
				DialogProfile(bot, ctx, u.ChatID(), nil, kbProfile)
			}
		})
		r.CallbackDataHandler("/bio/edit", func(u *router.Update) {
			DialogEditBio(bot, u.ChatID(), "/cancel")
			r.Expect(u, u.NeighborRoute("/bio/input"))
		})
		r.StateHandler("/bio/input", func(u *router.Update) {
			if u.Message == nil && u.Message.Text == "" {
				r.Expect(u, u.NeighborRoute("/bio/input"))
				RejectIncomingMessage(bot, u, "Попробуйте снова: бот не умеет обрабатывать такие сообщения, только текстовые")
				return
			}

			user := GetUser(u.Context())
			user.Bio = u.Message.Text
			if err := AddOrUpdateUser(u.Context(), user); err != nil {
				log.WithError(err).WithField("user-id", u.ChatID()).Error("Unable to save users bio to db")
				return
			}

			if !IsUserProfileCompleted(user) {
				r.Expect(u, u.NeighborRoute("/name/input"))
				DialogEditName(bot, u.ChatID(), "/cancel")
			} else {
				DialogProfile(bot, u.Context(), u.ChatID(), nil, kbProfile)
			}
		})
	})
	r.Route("/settings", func(r router.Router) {
		r.CallbackDataHandler("/show", func(u *router.Update) {
			DialogSettings(bot, u)
		})
		r.CallbackDataHandler("/visibility/{value}", func(u *router.Update) {
			err := SetSettingValue(u.Context(), "Visibility", router.GetParam(u.Context(), "value"))
			if err != nil {
				log.WithError(err).WithField("user-id", u.ChatID()).Error("Unable to set Visibility")
				return
			}
			DialogSettings(bot, u)
		})
		r.CallbackDataHandler("/visibility", func(u *router.Update) {
			DialogVisibility(bot, u.ChatID(), u)
		})
		r.CallbackDataHandler("/online/{value}", func(u *router.Update) {
			err := SetSettingValue(u.Context(), "OnLineReady", router.GetParam(u.Context(), "value"))
			if err != nil {
				log.WithError(err).WithField("user-id", u.ChatID()).Error("Unable to set OnLineReady")
				return
			}
			DialogSettings(bot, u)
		})
		r.CallbackDataHandler("/city/edit", func(u *router.Update) {
			SendMessage(bot, u.ChatID(), "Отправьте боту город (или несколько) где с вами можно встретиться, если хотите удалить информацию о городе отправьте <b>-</b> (минус)",
				WithKeyboard(
					MenuVertical(tgbotapi.NewInlineKeyboardButtonData("⬅️В меню", u.NeighborRoute("/show"))),
				),
			)
			r.Expect(u, u.NeighborRoute("/city/input"))
		})
		r.StateHandler("/city/input", func(u *router.Update) {
			if u.Message == nil && u.Message.Text == "" {
				r.Expect(u, u.NeighborRoute("/city/input"))
				RejectIncomingMessage(bot, u, "Попробуйте снова: бот не умеет обрабатывать такие сообщения, только текстовые")
				return
			}
			if u.Message.Text == "-" {
				err := DeleteSetting(u.Context(), "OfflineCity")
				if err != nil {
					log.WithError(err).WithField("user-id", u.ChatID()).Error("Unable to delete OfflineCity")
					return
				}
			} else {
				err := SetSettingValue(u.Context(), "OfflineCity", u.Message.Text)
				if err != nil {
					log.WithError(err).WithField("user-id", u.ChatID()).Error("Unable to set OfflineCity")
					return
				}
			}
			DialogSettings(bot, u)
		})

	})
	r.CommandHandler("/start", func(u *router.Update) {
		user := GetUser(u.Context())
		if user.Status == UserStatusNew {
			SendMessage(bot, u.ChatID(), "Вы не зарегистрированы, пожалуйста заполните профиль",
				WithKeyboard(
					MenuVertical(tgbotapi.NewInlineKeyboardButtonData("Перейти к профилю", "/profile/show")),
				),
				WithReplaceMessage(u),
			)
			return
		}
		SendMessage(bot, u.ChatID(), "Главное меню",
			WithKeyboard(
				MenuVertical(
					tgbotapi.NewInlineKeyboardButtonData("📝Профиль", "/profile/show"),
					tgbotapi.NewInlineKeyboardButtonData("⚙️Настройки", "/settings/show"),
					tgbotapi.NewInlineKeyboardButtonData("☕Попить кофе с сообщником!", "/match/search"),
				),
			),
			WithReplaceMessage(u),
		)
	})

	// Start polling Telegram for updates.
	updates := bot.GetUpdatesChan(updateConfig)
	for update := range updates {
		r.Serve(router.NewUpdate(update))
	}
}
