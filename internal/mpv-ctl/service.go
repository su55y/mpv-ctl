package mpvctl

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/blang/mpv"
)

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

var (
	rxBoolProperty = regexp.MustCompile(`^(t|f|true|false)$`)
	rxNumProperty  = regexp.MustCompile(`^(\d+)$`)
)

func NewService(socketPath string) *Service {
	stat, err := os.Stat(socketPath)
	if err != nil {
		log.Fatalf("check socket error: %v", err)
	}
	if stat.Mode()&os.ModeSocket == 0 {
		log.Fatalf("%#v not a socket file", socketPath)
	}
	return &Service{
		mpvc: mpv.NewClient(mpv.NewIPCClient(socketPath)),
	}
}

func (s *Service) LoadFile(path, flag string) error {
	if len(flag) == 0 {
		flag = mpv.LoadFileModeReplace
	}
	switch flag {
	case mpv.LoadFileModeReplace, mpv.LoadFileModeAppend, mpv.LoadFileModeAppendPlay:
		return s.mpvc.Loadfile(path, flag)
	}
	return fmt.Errorf("invalid 'flag' value: %s", flag)
}

func (s *Service) LoadList(path, flag string) error {
	if len(flag) == 0 {
		flag = mpv.LoadFileModeReplace
	}
	switch flag {
	case mpv.LoadFileModeReplace, mpv.LoadFileModeAppend:
		return s.mpvc.LoadList(path, flag)
	}
	return fmt.Errorf("invalid 'flag' value: %s", flag)
}

func (s *Service) Control(cmd string) error {
	switch controlType(cmd) {
	case Pause:
		return s.pause()
	case PauseCycle:
		return s.pauseCycle()
	case Play:
		return s.play()
	case Next:
		return s.next()
	case Prev:
		return s.prev()
	}
	return fmt.Errorf("cmd %+s not implemented", cmd)
}

func (s *Service) SetProperty(key, value string) error {
	prop, err := parsePropertyValue(value)
	if err != nil {
		return err
	}
	if err := s.mpvc.SetProperty(key, prop); err != nil {
		return err
	}
	return nil
}

func (s *Service) GetProperty(key string) (string, error) {
	return s.mpvc.GetProperty(key)
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
