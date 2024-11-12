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
	"golang.org/x/sys/windows/svc"
)

// myService for Windows service execution
type myService struct{}

func (m *myService) Execute(args []string, req <-chan svc.ChangeRequest, status chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	status <- svc.Status{State: svc.StartPending}
	status <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	run := true
	for run {
		select {
		case <-time.After(10 * time.Second):

		case c := <-req:
			switch c.Cmd {
			case svc.Stop, svc.Shutdown:
				run = false
				status <- svc.Status{State: svc.StopPending}
			}
		}
	}
	return false, 0
}

func main() {
	fmt.Printf("Log at : %s\n", utils.GetLogFileName())
	// Initialize the file lock
	// Try to acquire the lock
	// Ensure the lock file is released when the program exits
	// Will block here until user hits ctrl+c
	shouldReturn := run()
	if shouldReturn {
		return
	}
}

func run() bool {
	lockFileName := utils.GetLockFileName()
	fmt.Printf("Trying to get lock file from : %s\n", lockFileName)

	lock := utils.NewFileLock(lockFileName)

	if err := lock.Lock(); err != nil {
		fmt.Println(err)
		return true
	}
	fmt.Printf("Got lock file : %s\n", lockFileName)

	defer lock.Unlock()
	l := utils.GetLogger("")
	defer utils.CloseLogger()

	l.Error().Str("App name", utils.GetAppName()).Str("Working Dir", utils.GetAppRunningDir()).Send()
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Blocking, press ctrl+c to stop...")

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
	<-done
	fmt.Printf("\nClean shutdown : %s\n", utils.GetAppName())
	return false
}
