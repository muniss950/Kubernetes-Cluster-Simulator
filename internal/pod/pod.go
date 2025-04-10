package pod

import (
	"fmt"
	"github.com/google/uuid"
)

type Pod struct {
	ID     string `json:"id"`
	CPUs   int    `json:"cpus"`
	NodeID string `json:"node_id"` //ID of the node it is scheduled on
	Status string `json:"status"`  //e.g., Pending, Running, Failed
}

// CreatePod function to create a pod
func CreatePod(cpus int) Pod {
	podID := fmt.Sprintf("pod_%s", uuid.New().String())
	return Pod{
		ID:     podID,
		CPUs:   cpus,
		Status: "Pending", // Initial status
	}
}
