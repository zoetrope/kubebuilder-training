package controllers

import (
	"context"
	"time"

	multitenancyv1 "github.com/zoetrope/kubebuilder-training/codes/api/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func newExternalEventWatcher() *externalEventWatcher {
	ch := make(chan event.GenericEvent)

	return &externalEventWatcher{
		channel: ch,
	}
}

func (r *externalEventWatcher) InjectClient(c client.Client) error {
	r.client = c
	return nil
}

type externalEventWatcher struct {
	channel chan event.GenericEvent
	client  client.Client
}

func (r externalEventWatcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(10 * time.Second)

	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			var tenants multitenancyv1.TenantList
			err := r.client.List(ctx, &tenants, client.MatchingFields(map[string]string{conditionReadyField: string(corev1.ConditionTrue)}))
			if err != nil {
				break
			}
			for _, tenant := range tenants.Items {
				r.channel <- event.GenericEvent{
					Object: &tenant,
				}
			}
		}
	}
}
