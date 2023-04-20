package mpvctl

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/blang/mpv"
)

type controlType string

const (
	Pause      controlType = "pause"
	PauseCycle controlType = "pause-cycle"
	Play       controlType = "play"
	Next       controlType = "next"
	Prev       controlType = "prev"
)
const (
	MissedUrlErrMsg  = "'url' param should be included in query: %v"
	MissedCmdErrMsg  = "'cmd' param should be included in query: %v"
	InvalidCmdErrMsg = "invalid 'cmd' value: %v"
)

type LoaderParams struct {
	url  string
	flag string
}

type ControlParams struct {
	cmd controlType
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
	handleLoaderMethod(query, s.mpvc.Loadfile)(w)
}

func (s *Service) LoadPlaylist(query url.Values, w *http.ResponseWriter) {
	handleLoaderMethod(query, s.mpvc.LoadList)(w)
}

func (s *Service) Control(query url.Values, w *http.ResponseWriter) {
	s.handleControls(query)(w)
}

func handleLoaderMethod(
	query url.Values,
	callback func(string, string) error,
) func(*http.ResponseWriter) {
	params, err := parseLoaderParams(query)
	if err != nil {
		return func(w *http.ResponseWriter) { writeError(w, err.Error()) }
	}
	if err := callback(params.url, params.flag); err != nil {
		return func(w *http.ResponseWriter) { writeError(w, err.Error()) }
	}
	return func(w *http.ResponseWriter) { writeDefaultResponse(w) }
}

func (s *Service) handleControls(query url.Values) func(*http.ResponseWriter) {
	callback, err := s.parseControlParams(query)
	if err != nil {
		return func(w *http.ResponseWriter) { writeError(w, err.Error()) }
	}
	if err := callback(); err != nil {
		return func(w *http.ResponseWriter) { writeError(w, err.Error()) }
	}
	return func(w *http.ResponseWriter) { writeDefaultResponse(w) }
}

func parseLoaderParams(query url.Values) (*LoaderParams, error) {
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
	return &LoaderParams{query.Get("url"), flag}, nil
}

func (s *Service) parseControlParams(query url.Values) (func() error, error) {
	if !query.Has("cmd") {
		return nil, fmt.Errorf(MissedCmdErrMsg, query)
	}
	switch controlType(query.Get("cmd")) {
	case Pause:
		return s.pause, nil
	case PauseCycle:
		return s.pauseCycle, nil
	case Play:
		return s.play, nil
	case Next:
		return s.next, nil
	case Prev:
		return s.prev, nil
	default:
		return nil, fmt.Errorf(InvalidCmdErrMsg, query)
	}
}

func (s *Service) pauseCycle() error {
	pauseState, err := s.mpvc.Pause()
	if err != nil {
		return err
	}
	return s.mpvc.SetPause(!pauseState)
}

func (s *Service) play() error {
	return s.mpvc.SetPause(false)
}

func (s *Service) pause() error {
	return s.mpvc.SetPause(true)
}

func (s *Service) next() error {
	return s.mpvc.PlaylistNext()
}

func (s *Service) prev() error {
	return s.mpvc.PlaylistPrevious()
}
