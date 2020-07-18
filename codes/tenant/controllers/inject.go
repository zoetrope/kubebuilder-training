package controllers

import "context"

func (r *TenantReconciler) InjectStopChannel(ch <-chan struct{}) error {
	r.stopCh = ch
	return nil
}

func contextFromStopChannel(ch <-chan struct{}) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		<-ch
	}()
	return ctx
}
