package main

import (
	"flag"
	"log"
	"os"

	"github.com/blang/mpv"
	mpvctl "github.com/su55y/mpv-ctl/internal/mpv-ctl"
)

const (
	DEFAULT_SOCKET_PATH = "/tmp/mpv.sock"
)

var (
	cmd      string
	file     string
	playlist string
	loadflag string
	set      string
	value    string
	get      string

	client ICLIService
)

type ICLIService interface {
	ControlCliHandler(string) error
	LoadFileCliHandler(string, string) error
	LoadListCliHandler(string, string) error
	SetterCliHandler(string, string)
	GetterCliHandler(string)
}

func checkArgs() bool {
	for _, arg := range []string{cmd, file, playlist, get} {
		if len(arg) > 0 {
			return true
		}
	}
	if len(set) > 0 && len(value) > 0 {
		return true
	}
	return false
}

func parseArgs() {
	flag.StringVar(&cmd, "cmd", "", "cmd (play/pause/pause-cycle/next/prev)")
	flag.StringVar(&file, "video", "", "video path")
	flag.StringVar(&playlist, "playlist", "", "playlist path")
	flag.StringVar(&loadflag, "flag", "replace", "flag (append/append-play/replace)")
	flag.StringVar(&set, "set", "", "set property (property required)")
	flag.StringVar(&value, "value", "", "property value")
	flag.StringVar(&get, "get", "", "get property")
	flag.Parse()
	if !checkArgs() {
		flag.Usage()
		os.Exit(1)
	}
}

func initClient() {
	client = mpvctl.NewService(&mpv.Client{LLClient: mpv.NewIPCClient(DEFAULT_SOCKET_PATH)})
}

func init() {
	parseArgs()
	initClient()
}

func main() {
	var err error
	switch true {
	case len(cmd) > 0:
		err = client.ControlCliHandler(cmd)
	case len(file) > 0:
		err = client.LoadFileCliHandler(file, loadflag)
	case len(playlist) > 0:
		err = client.LoadListCliHandler(playlist, loadflag)
	case len(set) > 0:
		client.SetterCliHandler(set, value)
	case len(get) > 0:
		client.GetterCliHandler(get)
	default:
		flag.Usage()
		os.Exit(1)
	}
	if err != nil {
		log.Fatal(err)
	}
}
