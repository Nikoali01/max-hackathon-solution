package bot

import "context"

type Handler interface {
	Handle(ctx context.Context, req *Request, responder Responder) error
}

type HandlerFunc func(ctx context.Context, req *Request, responder Responder) error

func (f HandlerFunc) Handle(ctx context.Context, req *Request, responder Responder) error {
	return f(ctx, req, responder)
}
