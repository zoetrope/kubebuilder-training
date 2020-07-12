package controllers

import (
	"time"

	"sigs.k8s.io/controller-runtime/pkg/event"
)

func newExternalEventWatcher() *externalEventWatcher {
	ch := make(chan event.GenericEvent)

	return &externalEventWatcher{
		channel: ch,
	}
}

type externalEventWatcher struct {
	channel chan event.GenericEvent
}

func (r externalEventWatcher) Start(ch <-chan struct{}) error {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ch:
			return nil
		case <-ticker.C:
			//r.channel <- event.GenericEvent{
			//	Meta: &metav1.ObjectMeta{
			//		Name: "unknown-tenant",
			//	},
			//}
		}
	}
}
