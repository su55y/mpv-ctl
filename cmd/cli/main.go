package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/su55y/mpv-ctl/internal/service"
)

const (
	DEFAULT_SOCKET_PATH = "/tmp/mpv.sock"
)

var (
	cmd        string
	file       string
	playlist   string
	loadflag   string
	set        string
	value      string
	get        string
	socketPath string

	client ICLIService
)

type ICLIService interface {
	Load(string, string, bool) error
	Control(string) error
	SetProperty(string, string) error
	GetProperty(string) (string, error)
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
	flag.StringVar(&socketPath, "sock", DEFAULT_SOCKET_PATH, "socket path")
	flag.Parse()
	if !checkArgs() {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	parseArgs()
    var client ICLIService = service.NewService(socketPath)
	var err error
	switch true {
	case len(cmd) > 0:
		err = client.Control(cmd)
	case len(file) > 0:
		err = client.Load(file, loadflag, false)
	case len(playlist) > 0:
		err = client.Load(playlist, loadflag, true)
	case len(set) > 0:
		err = client.SetProperty(set, value)
	case len(get) > 0:
		var prop string
		if prop, err = client.GetProperty(get); len(prop) > 0 && err == nil {
			fmt.Print(prop)
		} else if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Printf("property %+s not found\n", get)
		}
	default:
		flag.Usage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
