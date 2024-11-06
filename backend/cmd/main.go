// Package main is the main entrypoint
package main

import (
	"github.com/vtpl1/phoring/backend/cmd/utils"
)

func main() {
	l := utils.GetLogger("")
	l.Error().Msgf("App Name: %s, working Dir: %s", utils.GetAppName(), utils.GetAppRunningDir())
}
