// Package main is the main entrypoint
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vtpl1/phoring/backend/cmd/utils"
	"github.com/vtpl1/phoring/backend/monitor"
)

func main() {
	l := utils.GetLogger("")
	l.Error().Msgf("App Name: %s, working Dir: %s", utils.GetAppName(), utils.GetAppRunningDir())
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Blocking, press ctrl+c to continue...")

	go func() {
		for {
			processName := "live_rec_service"
			metric, err := monitor.GetMetrics(processName)
			if err == nil {
				b, err1 := json.Marshal(metric)
				if err1 == nil {
					l.Info().Str("process", processName).Msg(string(b))
				}
			} else {
				l.Error().Str("process", processName).Err(err).Send()
			}
			time.Sleep(1 * time.Second)
		}
	}()
	<-done // Will block here until user hits ctrl+c
}
