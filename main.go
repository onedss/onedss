package main

import (
	"fmt"
	"github.com/onedss/EasyGoLib/utils"
	"github.com/onedss/onedss/app"
	"github.com/onedss/onedss/routers"
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
		log.SetFlags(log.Lshortfile | log.LstdFlags)
	}
	log.Println("git commit code :", gitCommitCode)
	log.Println("build date :", buildDateTime)
	routers.BuildVersion = fmt.Sprintf("%s.%s", routers.BuildVersion, gitCommitCode)
	routers.BuildDateTime = buildDateTime

	app.StartApp()
}
