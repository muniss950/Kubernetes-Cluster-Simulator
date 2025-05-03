package main

import (

    "os"
    "cluster-sim/api"
)

func main() {
	// Get port from environment variable or default to 8080
	port := "8080"
	if len(os.Args) > 1 {
		port = os.Args[1] // Use the first command-line argument as the port
	}

	api.StartServer(port)
}
