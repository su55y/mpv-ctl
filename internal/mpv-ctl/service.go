package mpvctl

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/blang/mpv"
)

type IHTTPService interface {
	LoadFileHttpHandler(url.Values, *http.ResponseWriter)
	ControlHttpHandler(url.Values, *http.ResponseWriter)
}

type ICLIService interface {
	ControlCliHandler(string) error
	LoadFileCliHandler(string, string) error
	LoadListCliHandler(string, string) error
	SetterCliHandler(string, string)
	GetterCliHandler(string)
}

type controlType string

type LoaderType struct {
	params   LoaderParams
	callback func(string, string) error
}

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

const (
	Pause      controlType = "pause"
	PauseCycle controlType = "pause-cycle"
	Play       controlType = "play"
	Next       controlType = "next"
	Prev       controlType = "prev"
)

const (
	MissedUrlErrMsg        = "'url' param should be included in query: %v"
	MissedCmdErrMsg        = "'cmd' param should be included in query: %v"
	InvalidFlagErrMsg      = "invalid 'flag' value: %s"
	InvalidCmdErrMsg       = "invalid 'cmd' value: %v"
	InvalidUrlQueryErrMsg  = "invalid url query: %v"
	PropertyNotFoundErrMsg = "'%s' propertyn not found"
)

var (
	rxBoolProperty = regexp.MustCompile(`^(t|f|true|false)$`)
	rxNumProperty  = regexp.MustCompile(`^(\d+)$`)
)

func NewService(client *mpv.Client) *Service {
	return &Service{
		mpvc: client,
	}
}

func (s *Service) LoadFileHttpHandler(query url.Values, w *http.ResponseWriter) {
	loader, err := s.parseLoadUrlQuery(query)
	if err != nil {
		writeError(w, err.Error())
		return
	}
	if err := loader.callback(loader.params.url, loader.params.flag); err != nil {
		writeError(w, err.Error())
	} else {
		writeDefaultResponse(w)
	}
}

func (s *Service) LoadFileCliHandler(path string, flag string) error {
	switch flag {
	case mpv.LoadFileModeReplace, mpv.LoadFileModeAppend, mpv.LoadFileModeAppendPlay:
		return s.mpvc.Loadfile(path, flag)
	}
	return fmt.Errorf(InvalidFlagErrMsg, flag)
}

func (s *Service) LoadListCliHandler(path string, flag string) error {
	switch flag {
	case mpv.LoadListModeReplace, mpv.LoadListModeAppend:
		return s.mpvc.LoadList(path, flag)
	}
	return fmt.Errorf(InvalidFlagErrMsg, flag)
}

func (s *Service) ControlHttpHandler(query url.Values, w *http.ResponseWriter) {
	callback, err := s.parseControlUrlQuery(query)
	if err != nil {
		writeError(w, err.Error())
		return
	}
	if err := callback(); err != nil {
		writeError(w, err.Error())
	} else {
		writeDefaultResponse(w)
	}
}

func (s *Service) SetterCliHandler(key string, value string) {
	val, err := parsePropertyValue(value)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := s.mpvc.SetProperty(key, val); err != nil {
		fmt.Println(err)
	}
}

func (s *Service) GetterCliHandler(key string) {
	value, err := s.mpvc.GetProperty(key)
	switch {
	case len(value) > 0 && err == nil:
		fmt.Print(value)
	case err != nil:
		fmt.Println(err)
	default:
		fmt.Printf(PropertyNotFoundErrMsg, key)
	}

}

func (s *Service) ControlCliHandler(cmd string) error {
	if callback := s.chooseControlCallback(controlType(cmd)); callback != nil {
		return callback()
	}
	return fmt.Errorf(InvalidCmdErrMsg, cmd)
}

func (s *Service) parseLoadUrlQuery(query url.Values) (*LoaderType, error) {
	if callback := s.chooseLoadCallback(query.Get("type")); callback != nil && query.Has("url") {
		url := query.Get("url")
		flag := mpv.LoadFileModeReplace
		switch query.Get("flag") {
		case mpv.LoadFileModeAppend:
			flag = mpv.LoadFileModeAppend
		case mpv.LoadFileModeAppendPlay:
			if flagtype := query.Get("type"); flagtype == "video" {
				flag = mpv.LoadFileModeAppendPlay
			}
		}
		return &LoaderType{LoaderParams{url, flag}, callback}, nil
	}
	return nil, fmt.Errorf(InvalidUrlQueryErrMsg, query)
}

func (s *Service) chooseLoadCallback(flagtype string) func(string, string) error {
	switch flagtype {
	case "video":
		return s.mpvc.Loadfile
	case "playlist":
		return s.mpvc.LoadList
	}
	return nil
}

func (s *Service) parseControlUrlQuery(query url.Values) (func() error, error) {
	if !query.Has("cmd") {
		return nil, fmt.Errorf(MissedCmdErrMsg, query)
	}
	if callback := s.chooseControlCallback(controlType(query.Get("cmd"))); callback != nil {
		return callback, nil
	}
	return nil, fmt.Errorf(InvalidCmdErrMsg, query)
}

func (s *Service) chooseControlCallback(ct controlType) func() error {
	switch ct {
	case Pause:
		return s.pause
	case PauseCycle:
		return s.pauseCycle
	case Play:
		return s.play
	case Next:
		return s.next
	case Prev:
		return s.prev
	}
	return nil
}

func parsePropertyValue(val string) (interface{}, error) {
	switch {
	case rxBoolProperty.MatchString(val):
		return strconv.ParseBool(val)
	case rxNumProperty.MatchString(val):
		return strconv.Atoi(val)
	}
	return val, nil
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
