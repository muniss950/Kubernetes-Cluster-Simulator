package health

import (
	"context"
	"log"
	"time"

	"cluster-sim/internal/node"
	// "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// HealthManager periodically checks the health of nodes.
type HealthManager struct {
	NodeManager *node.NodeManager
}

// NewHealthManager creates a new HealthManager.
func NewHealthManager(nm *node.NodeManager) *HealthManager {
	return &HealthManager{NodeManager: nm}
}

// StartMonitoring begins a goroutine that periodically inspects each node's container.
func (hm *HealthManager) StartMonitoring() {
	go func() {
		for {
			hm.checkNodesHealth()
			time.Sleep(10 * time.Second)
		}
	}()
}

// checkNodesHealth inspects the container for each node and updates its status.
func (hm *HealthManager) checkNodesHealth() {
	// Lock NodeManager to safely update the nodes map.

	hm.NodeManager.Mu.Lock()
	defer hm.NodeManager.Mu.Unlock()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Println("Error creating Docker client:", err)
		return
	}

	for id, node := range hm.NodeManager.Nodes {
		inspect, err := cli.ContainerInspect(context.Background(), id)
		if err != nil {
			log.Printf("Error inspecting container %s: %v", id, err)
			node.Status = "Unhealthy"

			if err := hm.NodeManager.RestartNode(id); err != nil {
				log.Printf("Auto-restart failed for node %s: %v", id, err)
			}

		} else {
			if inspect.State.Running {
				node.Status = "Running"
			} else {
				node.Status = "Stopped"
			}
		}
		hm.NodeManager.Nodes[id] = node
		log.Printf("Health Monitor: Node %s running", id)
	}
}
