package runners

import (
	"context"
	"fmt"
	"time"
)

type Runner struct {
}

func (r Runner) Start(ctx context.Context) error {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			fmt.Println("run something")
		}
	}
}

func (r Runner) NeedLeaderElection() bool {
	return true
}
