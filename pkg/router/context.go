package router

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Context struct {
	RoutePath  string
	ParentPath string
	Router     Router
	kv         map[string]string
}

type ContextKey string

var (
	ctxKeyRouteContext = ContextKey("tg-route-context")
	ctxKeyUpdate       = ContextKey("tg-update")
)

func RouteContext(ctx context.Context) *Context {
	rctx, _ := ctx.Value(ctxKeyRouteContext).(*Context)
	return rctx
}

func NewRouteContext() *Context {
	return &Context{
		kv: make(map[string]string),
	}
}

func (c *Context) AddParam(key, value string) {
	c.kv[key] = value
}

func (c *Context) Get(key string) string {
	if v, ok := c.kv[key]; !ok {
		return ""
	} else {
		return v
	}
}

type Update struct {
	tgbotapi.Update
	ctx    context.Context
	cancel context.CancelFunc
}

func NewUpdate(update tgbotapi.Update) *Update {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, ctxKeyUpdate, update)
	return &Update{
		Update: update,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (u *Update) ChatID() int64 {
	if u.Message != nil {
		return u.Message.From.ID
	} else {
		return u.CallbackQuery.From.ID
	}
}

func (u *Update) ParentRoutePath() string {
	return RouteContext(u.ctx).ParentPath
}

func (u *Update) RoutePath() string {
	return RouteContext(u.ctx).RoutePath
}

func (u *Update) NeighborRoute(t string) string {
	return fmt.Sprintf("%s%s", RouteContext(u.ctx).ParentPath, t)
}

func (u *Update) Context() context.Context {
	return u.ctx
}

func (u *Update) WithContext(ctx context.Context) *Update {
	if ctx == nil {
		panic("nil context")
	}
	u2 := new(Update)
	*u2 = *u
	u2.ctx = ctx
	return u2
}

func GetParam(ctx context.Context, name string) string {
	return RouteContext(ctx).Get("value")
}
