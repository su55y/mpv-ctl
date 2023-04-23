package main

import (
	"flag"
	"log"

	"github.com/blang/mpv"
	mpvctl "github.com/su55y/mpv-ctl/internal/mpv-ctl"
)

const (
	DEFAULT_SOCKET_PATH = "/tmp/mpv.sock"
)

var (
	cmd      string
	path     string
	loadflag string
	loadtype string

	client mpvctl.ICLIService
)

func parseArgs() {
	flag.StringVar(&cmd, "cmd", "", "cmd (play/pause/pause-cycle/next/prev)")
	flag.StringVar(&path, "path", "", "path")
	flag.StringVar(&loadflag, "flag", "replace", "flag (append/append-play/replace)")
	flag.StringVar(&loadtype, "type", "video", "(video/playlist)")
	flag.Parse()
	if len(cmd) == 0 && len(path) == 0 {
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

func runCmd() {
}

func main() {
	switch true {
	case len(cmd) > 0:
		if err := client.ControlCliHandler(cmd); err != nil {
			log.Fatal(err)
		}
	case len(path) > 0:
		if err := client.LoadFileCliHandler(path, loadflag, loadtype); err != nil {
			log.Fatal(err)
		}
	default:
		flag.Usage()
	}
}
