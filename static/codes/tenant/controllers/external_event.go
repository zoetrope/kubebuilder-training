package controllers

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func newExternalEvent() *externalEventSource {
	ch := make(chan event.GenericEvent)

	return &externalEventSource{
		channel: ch,
	}
}

type externalEventSource struct {
	channel chan event.GenericEvent
}

func (r externalEventSource) Start(ch <-chan struct{}) error {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ch:
			ticker.Stop()
			return nil
		case <-ticker.C:
			r.channel <- event.GenericEvent{
				Meta: &metav1.ObjectMeta{
					Name: "unknown-tenant",
				},
			}
		}
	}
}
