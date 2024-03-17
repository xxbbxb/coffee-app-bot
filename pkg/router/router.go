package router

// NewRouter returns a new Mux object that implements the Router interface.
func NewRouter(def HandlerFunc) *Mux {
	return NewMux(def)
}

// Router consisting of the core routing methods used by chi's Mux,
// using only the standard net/http.
type Router interface {
	Handler

	// Use appends one or more middlewares onto the Router stack.
	Use(middlewares ...func(Handler) Handler)
	// Same as above but inline
	With(middlewares ...func(Handler) Handler) Router
	// Route Mounts a sub-Router along a `pattern` string.
	Route(pattern string, fn func(r Router)) Router
	// Mount Handler for `pattern`
	CallbackDataHandler(pattern string, h HandlerFunc)
	// Mount Handler for command
	CommandHandler(pattern string, h HandlerFunc)
	// Mount state Handler for `pattern`
	StateHandler(pattern string, h HandlerFunc)
	// Set `pattern` for handling future input
	Expect(next *Update, pattern string)
	// Mount attaches another Handler along /pattern/*
	Mount(pattern string, h Handler)

	WithMiddlewares(h Handler) Handler
}
