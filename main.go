package main

import (
	"flag"
	"github.com/sysu-team/Back-end-development/app"
)

func main() {
	// todo: using command line option
	configFile := flag.String("c", "config.yaml", "Config file")
	flag.Parse()
	app.Run(*configFile)
}
