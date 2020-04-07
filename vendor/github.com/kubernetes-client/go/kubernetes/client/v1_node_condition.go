/*
 * Kubernetes
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: v1.10.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package client

import (
	"time"
)

// NodeCondition contains condition information for a node.
type V1NodeCondition struct {

	// Last time we got an update on a given condition.
	LastHeartbeatTime time.Time `json:"lastHeartbeatTime,omitempty"`

	// Last time the condition transit from one status to another.
	LastTransitionTime time.Time `json:"lastTransitionTime,omitempty"`

	// Human readable message indicating details about last transition.
	Message string `json:"message,omitempty"`

	// (brief) reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`

	// Status of the condition, one of True, False, Unknown.
	Status string `json:"status"`

	// Type of node condition.
	Type_ string `json:"type"`
}
