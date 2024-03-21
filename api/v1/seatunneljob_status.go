package v1

import (
	corev1 "k8s.io/api/core/v1"
	"time"
)

type ClusterConditionType string

const (
	ClusterConditionJobSucceeded ClusterConditionType = "JobSucceeded"
	ClusterConditionError                             = "Error"
)

// MembersStatus is the status of the members of the cluster with both
// ready and unready node membership lists
type MembersStatus struct {
	//+nullable
	Succeeded []string `json:"succeeded,omitempty"`
	//+nullable
	Unsucceeded []string `json:"unsucceeded,omitempty"`
	//+nullable
	Failed []string `json:"failed,omitempty"`
}

// ClusterCondition shows the current condition of a cluster.
// Comply with k8s API conventions
type ClusterCondition struct {
	// Type of cluster condition.
	Type ClusterConditionType `json:"type,omitempty"`

	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status,omitempty"`

	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`

	// A human-readable message indicating details about the transition.
	Message string `json:"message,omitempty"`

	// The last time this condition was updated.
	LastUpdateTime string `json:"lastUpdateTime,omitempty"`

	// Last time the condition transitioned from one status to another.
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
}

// SeatunnelJobStatus defines the observed state of the seatunneljob
type SeatunnelJobStatus struct {
	// Members is the members in the cluster
	Members MembersStatus `json:"members,omitempty"`

	// Replicas is the number of desired replicas in the cluster
	Replicas int32 `json:"replicas,omitempty"`

	// SucceededReplicas is the number of succeeded replicas in the cluster
	SucceededReplicas int32 `json:"succeededReplicas,omitempty"`

	// Completed is status of the seatunneljob
	Completed bool `json:"completed"`

	// Conditions list all the applied conditions
	Conditions []ClusterCondition `json:"conditions,omitempty"`
}

func (zs *SeatunnelJobStatus) Init() {
	// Initialise conditions
	conditionTypes := []ClusterConditionType{
		ClusterConditionJobSucceeded,
		ClusterConditionError,
	}
	for _, conditionType := range conditionTypes {
		if _, condition := zs.GetClusterCondition(conditionType); condition == nil {
			c := newClusterCondition(conditionType, corev1.ConditionFalse, "", "")
			zs.setClusterCondition(*c)
		}
	}
}

func newClusterCondition(condType ClusterConditionType, status corev1.ConditionStatus, reason, message string) *ClusterCondition {
	return &ClusterCondition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastUpdateTime:     "",
		LastTransitionTime: "",
	}
}

func (zs *SeatunnelJobStatus) SetJobSucceededConditionTrue() {
	c := newClusterCondition(ClusterConditionJobSucceeded, corev1.ConditionTrue, "", "")
	zs.setClusterCondition(*c)
}

func (zs *SeatunnelJobStatus) SetJobSucceededConditionFalse() {
	c := newClusterCondition(ClusterConditionJobSucceeded, corev1.ConditionFalse, "", "")
	zs.setClusterCondition(*c)
}

func (zs *SeatunnelJobStatus) SetErrorConditionTrue(reason, message string) {
	c := newClusterCondition(ClusterConditionError, corev1.ConditionTrue, reason, message)
	zs.setClusterCondition(*c)
}

func (zs *SeatunnelJobStatus) SetErrorConditionFalse() {
	c := newClusterCondition(ClusterConditionError, corev1.ConditionFalse, "", "")
	zs.setClusterCondition(*c)
}

func (zs *SeatunnelJobStatus) GetClusterCondition(t ClusterConditionType) (int, *ClusterCondition) {
	for i, c := range zs.Conditions {
		if t == c.Type {
			return i, &c
		}
	}
	return -1, nil
}

func (zs *SeatunnelJobStatus) setClusterCondition(newCondition ClusterCondition) {
	now := time.Now().Format(time.RFC3339)
	position, existingCondition := zs.GetClusterCondition(newCondition.Type)

	if existingCondition == nil {
		zs.Conditions = append(zs.Conditions, newCondition)
		return
	}

	if existingCondition.Status != newCondition.Status {
		existingCondition.Status = newCondition.Status
		existingCondition.LastTransitionTime = now
		existingCondition.LastUpdateTime = now
	}

	if existingCondition.Reason != newCondition.Reason || existingCondition.Message != newCondition.Message {
		existingCondition.Reason = newCondition.Reason
		existingCondition.Message = newCondition.Message
		existingCondition.LastUpdateTime = now
	}

	zs.Conditions[position] = *existingCondition
}

func (zs *SeatunnelJobStatus) IsClusterInSucceededState() bool {
	_, readyCondition := zs.GetClusterCondition(ClusterConditionJobSucceeded)
	if readyCondition != nil && readyCondition.Status == corev1.ConditionTrue {
		return true
	}
	return false
}
