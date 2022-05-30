package main

import (
	"fmt"
	"github.com/onedss/onedss/app"
	"github.com/onedss/onedss/routers"
	"log"
)

var (
	gitCommitCode string
	buildDateTime string
)

func main() {
	log.Println("git commit code :", gitCommitCode)
	log.Println("build date :", buildDateTime)
	routers.BuildVersion = fmt.Sprintf("%s.%s", routers.BuildVersion, gitCommitCode)
	routers.BuildDateTime = buildDateTime

	app.StartApp()
}
