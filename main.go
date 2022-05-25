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
	log.Printf("git commit code:%s", gitCommitCode)
	log.Printf("build date:%s", buildDateTime)
	routers.BuildVersion = fmt.Sprintf("%s.%s", routers.BuildVersion, gitCommitCode)
	routers.BuildDateTime = buildDateTime

	app.StartApp()
}
