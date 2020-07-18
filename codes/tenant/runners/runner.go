package runners

import (
	"fmt"
	"time"
)

type Runner struct {
}

func (r Runner) Start(ch <-chan struct{}) error {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ch:
			return nil
		case <-ticker.C:
			fmt.Println("run something")
		}
	}
}

func (r Runner) NeedLeaderElection() bool {
	return true
}
