package main

import (
	"time"

	"github.com/go-logr/logr"
)

type runner struct {
	log logr.Logger
}

func (r runner) Start(ch <-chan struct{}) error {
	timer := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ch:
			return nil
		case <-timer.C:
			r.log.Info("run something")
		}
	}
}

func (r runner) NeedLeaderElection() bool {
	return true
}
