package router

import (
	"context"
	"fmt"
)

type Mux struct {
	tree        *node
	handler     Handler
	middlewares Middlewares
	state       *State
	entrypoint  *Mux
}

func NewMux(def Handler) *Mux {
	m := Mux{
		tree:        &node{},
		state:       &State{},
		handler:     def,
		middlewares: make(Middlewares, 0),
	}
	return &m
}

func (mx *Mux) Default(h Handler) {
	mx.handler = h
}

func (mx *Mux) Serve(u *Update) {
	if mx.handler == nil {
		panic("mux: no default handler")
	}
	var h Handler
	rctx, _ := u.Context().Value(ctxKeyRouteContext).(*Context)
	if rctx != nil {
		if mx.handler == nil {
			panic("mux: subrouter must have at least one of CallbackData or Command")
		}
		mx.handler.Serve(u)
		return
	}

	rctx = NewRouteContext()
	expect := mx.state.GetState(u.ChatID())
	// Order matters
	switch {
	case u.CallbackQuery != nil:
		h = mx.tree.route(rctx, u.CallbackQuery.Data)
		mx.state.Clear(u.ChatID())
	case expect != "":
		h = mx.tree.route(rctx, expect)
		mx.state.Clear(u.ChatID())
	case u.Message != nil:
		h = mx.tree.route(rctx, u.Message.Text)
	}
	if h == nil {
		mx.WithMiddlewares(mx.handler).Serve(u)
		return
	}
	u = u.WithContext(context.WithValue(u.Context(), ctxKeyRouteContext, rctx))

	if rctx.Router != nil {
		rctx.Router.WithMiddlewares(h).Serve(u)
	} else {
		mx.WithMiddlewares(h).Serve(u)
	}
}
func (mx *Mux) Use(mws ...func(Handler) Handler) {
	mx.middlewares = append(mx.middlewares, mws...)
}

func (mx *Mux) With(mws ...func(Handler) Handler) Router {
	mx.middlewares = append(mx.middlewares, mws...)
	return mx
}

func (mx *Mux) Route(pattern string, fn func(r Router)) Router {
	if fn == nil {
		panic(fmt.Sprintf("mux: attempting to Route() a nil subrouter on '%s'", pattern))
	}
	subRouter := NewRouter(nil)
	subRouter.middlewares = append(subRouter.middlewares, mx.middlewares...)
	mx.Mount(pattern, subRouter)
	fn(subRouter)
	/*if subRouter.handler == nil {
		subRouter.handler = mx.handler
	}*/
	return subRouter
}

func (mx *Mux) CallbackDataHandler(pattern string, h HandlerFunc) {
	mx.Mount(pattern, h)
}

func (mx *Mux) Expect(u *Update, pattern string) {
	if mx.entrypoint == nil {
		mx.state.SetState(u.ChatID(), pattern)
		return
	}
	mx.entrypoint.Expect(u, pattern)
}

func (mx *Mux) ExpectClean(u *Update) {
	if mx.entrypoint == nil {
		mx.state.Clear(u.ChatID())
		return
	}
	mx.entrypoint.ExpectClean(u)
}

func (mx *Mux) StateHandler(pattern string, h HandlerFunc) {
	mx.Mount(pattern, h)
}

func (mx *Mux) CommandHandler(pattern string, h HandlerFunc) {
	mx.Mount(pattern, h)
}

func (mx *Mux) Mount(pattern string, handler Handler) {
	if handler == nil {
		panic(fmt.Sprintf("mux: attempting to Mount() a nil handler on '%s'", pattern))
	}
	n := mx.tree.addChild(pattern, handler)
	if mux, ok := handler.(*Mux); ok {
		mux.handler = mx.handler // attach default handler
		mux.tree = n
		if mx.entrypoint == nil {
			mux.entrypoint = mx
		} else {
			mux.entrypoint = mx.entrypoint
		}
		mux.middlewares = mx.middlewares
	}
}

func (mx *Mux) WithMiddlewares(h Handler) Handler {
	return mx.middlewares.Handler(h)
}
