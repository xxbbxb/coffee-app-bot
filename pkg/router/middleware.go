package router

type Handler interface {
	Serve(*Update)
}

type HandlerFunc func(*Update)

func (f HandlerFunc) Serve(u *Update) {
	f(u)
}

type Middlewares []func(Handler) Handler

func (mws Middlewares) Handler(endpoint Handler) Handler {
	return &ChainHandler{endpoint, mws}
}

func (mws Middlewares) HandlerFunc(endpoint HandlerFunc) Handler {
	return &ChainHandler{endpoint, mws}
}

type ChainHandler struct {
	Endpoint Handler
	mws      Middlewares
}

func (c *ChainHandler) Serve(u *Update) {
	if len(c.mws) == 0 {
		c.Endpoint.Serve(u)
		return
	}
	h := c.mws[len(c.mws)-1](c.Endpoint)
	for i := len(c.mws) - 2; i >= 0; i-- {
		h = c.mws[i](h)
	}
	h.Serve(u)
}
