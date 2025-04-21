package node
import (
	"cluster-sim/internal/pod"
	"context"
	"log"
	"sync"
	"time"
  "fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)
// NodeManager manages the nodes in the cluster
type NodeManager struct {
    Nodes map[string]Node
    Pods map[string]pod.Pod
    Mu    sync.Mutex // Protects concurrent access to the nodes map
    totalCPUs int //Simulate resource pool
}

// NewNodeManager creates a new NodeManager
func NewNodeManager() *NodeManager {
    return &NodeManager{
        Nodes: make(map[string]Node),
        Pods:  make(map[string]pod.Pod),
        totalCPUs: 0,
    }
}

// AddNode adds a node to the cluster
func (nm *NodeManager) AddNode(node Node) {
    nm.Mu.Lock()
    defer nm.Mu.Unlock()
    nm.Nodes[node.ID] = node
    nm.totalCPUs += node.CPUs // Simulate resource allocation
}

// GetNodes returns all nodes in the cluster
func (nm *NodeManager) GetNodes() map[string]Node {
    nm.Mu.Lock()
    defer nm.Mu.Unlock()
    return nm.Nodes
}

// checkNodeHealth checks if a node's container is running
func checkNodeHealth(containerID string) (bool, error) {
    cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        return false, err
    }

    inspect, err := cli.ContainerInspect(context.Background(), containerID)
    if err != nil {
        return false, err
    }

    return inspect.State.Running, nil
}


func (nm *NodeManager) RestartNode(nodeID string) error {
    // nm.Mu.Lock()
    _, exists := nm.Nodes[nodeID]
    log.Print("stuff")
    // defer nm.Mu.Unlock()
    if !exists {
        return fmt.Errorf("node not found")
    }
    log.Print("here")

    cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        return err
    }

    if err :=RestartNodeContainer(nodeID);  err != nil {
        return err
    }
    log.Print("done")

    log.Printf("Node %s restarted", nodeID)
    time.Sleep(5 * time.Second)

    healthy, err := checkNodeHealth(nodeID)
    if err != nil || !healthy {
        log.Printf("Node %s still unhealthy, removing and rescheduling...", nodeID)
        _ = cli.ContainerRemove(context.Background(), nodeID, container.RemoveOptions{Force: true})
        nm.reschedulePods(nodeID)
        return fmt.Errorf("node restart failed and was removed")
    }

    return nil
}

func (nm *NodeManager) ShutdownNodes() {
	nm.Mu.Lock()
	defer nm.Mu.Unlock()

	log.Println("Stopping all nodes...")

	for id := range nm.Nodes {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			log.Printf("Error creating Docker client for node %s: %v", id, err)
			continue
		}

		if err := cli.ContainerStop(context.Background(), id, container.StopOptions{}); err != nil {
			log.Printf("Error stopping node %s: %v", id, err)
		} else {
			log.Printf("Node %s stopped", id)
		}
	}

	log.Println("All nodes have been stopped.")
}
