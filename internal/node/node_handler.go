// All the gin handlers are here for node package
package node

import (
	"cluster-sim/internal/pod"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
  "os"
  "os/signal"
  "syscall"
  "context"
)

// API Handler to add a new node
func (nm *NodeManager) AddNodeHandler(c *gin.Context) {
	var request struct {
		CPUs int `json:"cpus"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	id, err := CreateNodeContainer(request.CPUs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	newNode := Node{
		ID:        id,
		CPUs:      request.CPUs,
		UsedCPUs:  0,
		Status:    "Running",
		Pods:      []string{},
		CreatedAt: time.Now(),
	}
	nm.AddNode(newNode)
	log.Printf("Node created: id=%s, cpus=%d", id, request.CPUs)
	//Simulate Heartbeat Initialization
	println("Simulating heartbeat initialization for node:", id)

	c.JSON(http.StatusOK, gin.H{"message": "Node added", "node_id": id})
}

// API Handler to list all nodes with health status
func (nm *NodeManager) ListNodesHandler(c *gin.Context) {
	responseNodes := make([]gin.H, 0, len(nm.Nodes))
	nodes := nm.GetNodes()
	for _, node := range nodes {
		// Check node health
		healthy, err := checkNodeHealth(node.ID)
		if err != nil {
			// Log the error but continue
			println("Error checking node health:", err.Error())
			node.Status = "Unhealthy" // Or "Error"
		} else if healthy {
			node.Status = "Running"
		} else {
			node.Status = "Stopped"
		}
		//log each node details
		log.Printf("Listing node: id=%s, cpus=%d, used_cpus=%d, status=%s", node.ID, node.CPUs, node.UsedCPUs, node.Status)
		responseNodes = append(responseNodes, gin.H{
			"id":        node.ID,
			"cpus":      node.CPUs,
			"used_cpus": node.UsedCPUs,
			"status":    node.Status,
			"pods":      node.Pods,
		})
	}

	c.JSON(http.StatusOK, responseNodes)
}

// API Handler to add a new pod
func (nm *NodeManager) AddPodHandler(c *gin.Context) {
	var request struct {
		CPUs      int    `json:"cpus"`
		Algorithm string `json:"algorithm"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	nm.Mu.Lock()
	defer nm.Mu.Unlock()

	// Create a pod
	newPod := pod.CreatePod(request.CPUs)
	log.Printf("Pod created (pending): id=%s, cpus=%d", newPod.ID, request.CPUs)

	// Schedule the pod
	nodeID, err := SchedulePod(newPod, nm.Nodes, request.Algorithm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update the pod's NodeID
	newPod.NodeID = nodeID
	newPod.Status = "Running" // Update pod status
	log.Printf("Pod scheduled: pod_id=%s, assigned_node=%s", newPod.ID, nodeID)
	nm.Pods[newPod.ID] = newPod
	c.JSON(http.StatusOK, gin.H{"message": "Pod scheduled", "node_id": nodeID, "pod_id": newPod.ID})
}

func (nm *NodeManager) RestartNodeHandler(c *gin.Context) {
	var request struct {
		NodeID string `json:"node_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := nm.RestartNode(request.NodeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Node restarted successfully", "node_id": request.NodeID})
}

func (nm *NodeManager) DeleteNodeHandler(c *gin.Context) {
	var request struct {
		NodeID string `json:"node_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	// Create a Docker client.
	// Remove the container.
	if err := DeleteNodeContainer(request.NodeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("Docker container %s removed", request.NodeID)

	nm.Mu.Lock()
	nodeObj, exists := nm.Nodes[request.NodeID]
	if !exists {
		nm.Mu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}
	delete(nm.Nodes, request.NodeID)
	nm.totalCPUs -= nodeObj.CPUs
	nm.Mu.Unlock()

	log.Printf("Node %s deleted", request.NodeID)
	nm.reschedulePods(request.NodeID)
	c.JSON(http.StatusOK, gin.H{"message": "Node deleted and pods rescheduled", "node_id": request.NodeID})
}
func (nm *NodeManager) ShutdownHandler(srv *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit // Wait for shutdown signal

	log.Println("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Gracefully stop the HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Shutdown all nodes before exiting
	log.Println("Shutting down nodes...")
	nm.ShutdownNodes()

	log.Println("Server exited cleanly.")
}
