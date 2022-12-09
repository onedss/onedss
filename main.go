package main

import (
	"fmt"
	"github.com/onedss/onedss/app"
	"github.com/onedss/onedss/buildtime"
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
	log.SetPrefix("[OneDss] ")
	log.SetFlags(log.LstdFlags)
	if utils.Debug {
		log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds)
	}
	routers.BuildVersion = fmt.Sprintf("%s.%s", routers.BuildVersion, gitCommitCode)
	routers.BuildDateTime = fmt.Sprintf("<%s> %s", buildtime.BuildTime.Format(utils.DateTimeLayout), buildDateTime)
	core.Info("BuildVersion:", routers.BuildVersion)
	core.Info("BuildTime:", routers.BuildDateTime)

	app.StartApp()
}
