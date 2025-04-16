package api

import (
	"cluster-sim/internal/health"
	"cluster-sim/internal/node"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func StartServer(port string) {
	r := gin.Default()

	// Initialize NodeManager
	nodeManager := node.NewNodeManager()

	// Initialize Health Manager
	healthManager := health.NewHealthManager(nodeManager)
	healthManager.StartMonitoring()

	// Register routes, binding the NodeManager
	r.POST("/add_node", nodeManager.AddNodeHandler)
	r.GET("/nodes", nodeManager.ListNodesHandler)
	r.POST("/add_pod", nodeManager.AddPodHandler) // Added this line
	r.PUT("/restart_node", nodeManager.RestartNodeHandler)
	r.DELETE("/delete_node", nodeManager.DeleteNodeHandler)

	// log.Printf("API Server running on port %s\n", port)
	// r.Run(":" + port)
	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Run the server in a separate goroutine
	go func() {
		log.Printf("API Server running on port %s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Handle graceful shutdown
	nodeManager.ShutdownHandler(srv)
}
