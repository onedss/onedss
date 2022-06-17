package main

import (
	"fmt"
	"github.com/onedss/onedss/app"
	"github.com/onedss/onedss/core"
	"github.com/onedss/onedss/routers"
	"github.com/onedss/onedss/utils"
	"log"
)

var (
	gitCommitCode string
	buildDateTime string
)

func main() {
	log.SetPrefix("[EasyDarwin] ")
	log.SetFlags(log.LstdFlags)
	if utils.Debug {
		log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)
	}
	core.Info("git commit code :", gitCommitCode)
	core.Info("build date :", buildDateTime)
	routers.BuildVersion = fmt.Sprintf("%s.%s", routers.BuildVersion, gitCommitCode)
	routers.BuildDateTime = buildDateTime

	app.StartApp()
}
