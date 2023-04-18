package mpvctl

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/blang/mpv"
)

type methodType uint8

const (
	AppendVideo = iota
	LoadPlaylist
)
const (
	MissedUrlErrMsg = "'url' param should be present: (%+v)"
)

type ParamsType1 struct {
	url  string
	flag string
}

type Service struct {
	mpvc *mpv.Client
}

func NewService(client *mpv.Client) *Service {
	return &Service{
		mpvc: client,
	}
}

func (s *Service) LoadFile(query url.Values, w *http.ResponseWriter) {
	handleMethodType1(query, s.mpvc.Loadfile)(w)
}

func (s *Service) LoadPlaylist(query url.Values, w *http.ResponseWriter) {
	handleMethodType1(query, s.mpvc.LoadList)(w)
}

func handleMethodType1(
	query url.Values,
	callback func(string, string) error,
) func(*http.ResponseWriter) {
	params, err := parseParamsType1(query)
	if err != nil {
		return func(w *http.ResponseWriter) { writeError(w, err.Error()) }
	}
	if err := callback(params.url, params.flag); err != nil {
		return func(w *http.ResponseWriter) { writeError(w, err.Error()) }
	}
	return func(w *http.ResponseWriter) { writeDefaultResponse(w) }
}

func parseParamsType1(query url.Values) (*ParamsType1, error) {
	if !query.Has("url") {
		return nil, fmt.Errorf(MissedUrlErrMsg, query)
	}
	flag := mpv.LoadFileModeReplace
	if query.Has("flag") {
		switch query.Get("flag") {
		case mpv.LoadFileModeAppend:
			flag = mpv.LoadFileModeAppend
		case mpv.LoadFileModeAppendPlay:
			flag = mpv.LoadFileModeAppendPlay
		}
	}
	return &ParamsType1{query.Get("url"), flag}, nil
}
