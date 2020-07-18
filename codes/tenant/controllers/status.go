package controllers

import (
	"time"

	multitenancyv1 "github.com/zoetrope/kubebuilder-training/codes/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func setCondition(conditions *[]multitenancyv1.TenantCondition, newCondition multitenancyv1.TenantCondition) {
	if conditions == nil {
		conditions = &[]multitenancyv1.TenantCondition{}
	}
	current := findCondition(*conditions, newCondition.Type)
	if current == nil {
		newCondition.LastTransitionTime = metav1.NewTime(time.Now())
		*conditions = append(*conditions, newCondition)
		return
	}
	if current.Status != newCondition.Status {
		current.Status = newCondition.Status
		current.LastTransitionTime = metav1.NewTime(time.Now())
	}
	current.Reason = newCondition.Reason
	current.Message = newCondition.Message
}

func findCondition(conditions []multitenancyv1.TenantCondition, conditionType multitenancyv1.TenantConditionType) *multitenancyv1.TenantCondition {
	for _, c := range conditions {
		if c.Type == conditionType {
			return &c
		}
	}
	return nil
}
