package bot

type Chain struct {
	middlewares []Middleware
}

func NewChain(middlewares ...Middleware) Chain {
	return Chain{append(([]Middleware)(nil), middlewares...)}
}

func (c Chain) Then(h Handler) Handler {
	for i := range c.middlewares {
		h = c.middlewares[len(c.middlewares)-1-i](h)
	}
	return h
}

func (c Chain) ThenFunc(hf HandleFunc) Handler {
	return c.Then(hf)
}

func (c Chain) Append(middlewares ...Middleware) Chain {
	newCons := make([]Middleware, 0, len(c.middlewares)+len(middlewares))
	newCons = append(newCons, c.middlewares...)
	newCons = append(newCons, middlewares...)

	return Chain{newCons}
}

func (c Chain) Extend(nc Chain) Chain {
	return c.Append(nc.middlewares...)
}
