package protolizer

import (
	"context"
	"net/url"
)

type Header map[string][]string
type Options map[string]string

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
	options any
}

func UnwrapRpcOptions[T any](rpcOption *RpcOption) (*T, bool) {
	out, ok := rpcOption.options.(T)
	return &out, ok
}

type Server interface {
	Handler(RpcOption, func(context.Context, *Request[Reflected]) (*Response[Reflected], error))
}
