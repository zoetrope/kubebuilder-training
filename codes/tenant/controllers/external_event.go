package controllers

import (
	"time"

	multitenancyv1 "github.com/zoetrope/kubebuilder-training/codes/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (r externalEventWatcher) Start(ch <-chan struct{}) error {
	ticker := time.NewTicker(10 * time.Second)
	ctx := contextFromStopChannel(ch)

	defer ticker.Stop()
	for {
		select {
		case <-ch:
			return nil
		case <-ticker.C:
			var tenants multitenancyv1.TenantList
			err := r.client.List(ctx, &tenants, client.MatchingFields(map[string]string{conditionReadyField: string(corev1.ConditionTrue)}))
			if err != nil {
				break
			}
			for _, tenant := range tenants.Items {
				r.channel <- event.GenericEvent{
					Meta: &metav1.ObjectMeta{
						Name: tenant.Name,
					},
				}
			}
		}
	}
}
