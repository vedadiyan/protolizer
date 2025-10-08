package protolizer

import (
	"context"
	"net/url"
)

type Header map[string][]string
type Options map[string]string
type Handler[I Reflected, O Reflected] func(*Request[I]) (*Response[O], error)

type Request[T Reflected] struct {
	Headers    Header
	Data       T
	Url        *url.URL
	RemoteAddr string
	RemoteUri  string
	Options    Options
}

type Response[T Reflected] struct {
	Headers    Header
	Data       T
	StatusCode int
	Options    Options
}

type RpcOption struct {
	Options Options
}

type Server interface {
	Handler(RpcOption, func(context.Context, *Request[Reflected]) (*Response[Reflected], error))
}
