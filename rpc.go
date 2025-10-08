package protolizer

import (
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
