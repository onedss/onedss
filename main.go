package main

import (
	"fmt"
	"github.com/onedss/EasyGoLib/utils"
	"github.com/onedss/onedss/app"
	"github.com/onedss/onedss/core/logger"
	"github.com/onedss/onedss/routers"
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
		log.SetFlags(log.Lshortfile | log.LstdFlags)
	}
	logger.Info("git commit code :", gitCommitCode)
	logger.Info("build date :", buildDateTime)
	routers.BuildVersion = fmt.Sprintf("%s.%s", routers.BuildVersion, gitCommitCode)
	routers.BuildDateTime = buildDateTime

	app.StartApp()
}
