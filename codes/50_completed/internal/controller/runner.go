package controller

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	viewv1 "github.com/zoetrope/markdown-view/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type Runner struct {
	client   client.Client
	logger   logr.Logger
	interval time.Duration
	channel  chan<- event.TypedGenericEvent[*viewv1.MarkdownView]
}

func NewRunner(client client.Client, logger logr.Logger, interval time.Duration, channel chan<- event.TypedGenericEvent[*viewv1.MarkdownView]) *Runner {
	return &Runner{
		client:   client,
		logger:   logger,
		interval: interval,
		channel:  channel,
	}
}

func (r Runner) Start(ctx context.Context) error {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			r.notify(ctx)
		}
	}
}

func (r Runner) notify(ctx context.Context) {
	var mdviewList viewv1.MarkdownViewList
	err := r.client.List(ctx, &mdviewList)
	if err != nil {
		r.logger.Error(err, "failed to list MarkdownView")
		return
	}

	for _, sts := range mdviewList.Items {
		r.channel <- event.TypedGenericEvent[*viewv1.MarkdownView]{
			Object: sts.DeepCopy(),
		}
	}
}

func (r Runner) NeedLeaderElection() bool {
	return true
}
