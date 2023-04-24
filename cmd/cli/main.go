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

	client mpvctl.ICLIService
)

func parseArgs() {
	flag.StringVar(&cmd, "cmd", "", "cmd (play/pause/pause-cycle/next/prev)")
	flag.StringVar(&file, "video", "", "video path")
	flag.StringVar(&playlist, "playlist", "", "playlist path")
	flag.StringVar(&loadflag, "flag", "replace", "flag (append/append-play/replace)")
	flag.Parse()
	if len(cmd) == 0 && len(file) == 0 && len(playlist) == 0 {
		flag.Usage()
		log.Fatal("cmd or path are required")
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
	default:
		flag.Usage()
		os.Exit(1)
	}
	if err != nil {
		log.Fatal(err)
	}
}
