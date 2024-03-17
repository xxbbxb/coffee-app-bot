package router

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Maybe(mw func(Handler) Handler, maybeFn func(u *Update) bool) func(Handler) Handler {
	return func(next Handler) Handler {
		return HandlerFunc(func(u *Update) {
			if maybeFn(u) {
				mw(next).Serve(u)
			} else {
				next.Serve(u)
			}
		})
	}
}

func When(h Handler, maybeFn func(u *Update) bool) func(Handler) Handler {
	return func(next Handler) Handler {
		return HandlerFunc(func(u *Update) {
			if maybeFn(u) {
				h.Serve(u)
			} else {
				next.Serve(u)
			}
		})
	}
}
func AutoCallbackOK(bot *tgbotapi.BotAPI) func(next Handler) Handler {
	return func(next Handler) Handler {
		return HandlerFunc(func(u *Update) {
			next.Serve(u)
			if u.CallbackQuery != nil {
				cb := tgbotapi.NewCallback(u.CallbackQuery.ID, "ok")
				bot.Send(cb)
			}
		})
	}
}

func New(h Handler) func(next Handler) Handler {
	return func(next Handler) Handler {
		return HandlerFunc(func(u *Update) {
			h.Serve(u)
		})
	}
}

func One() func(next Handler) Handler {
	return func(next Handler) Handler {
		return HandlerFunc(func(u *Update) {
			fmt.Println("Hello from one")
			next.Serve(u)
		})
	}
}

func Two() func(next Handler) Handler {
	return func(next Handler) Handler {
		return HandlerFunc(func(u *Update) {
			fmt.Println("Hello from two")
			next.Serve(u)
		})
	}
}

func Three() func(next Handler) Handler {
	return func(next Handler) Handler {
		return HandlerFunc(func(u *Update) {
			fmt.Println("Hello from three")
			next.Serve(u)
		})
	}
}
