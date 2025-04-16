package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "strings"

    "github.com/urfave/cli/v2"
)

type Node struct {
    ID       string   `json:"id"`
    CPUs     int      `json:"cpus"`
    UsedCPUs int      `json:"used_cpus"`
    Status   string   `json:"status"`
    Pods     []string `json:"pods"`
}

type NodeRequest struct {
    CPUs int `json:"cpus"`
}

type DeleteNodeRequest struct {
    NodeID string `json:"node_id"`
}

type RestartNodeRequest struct {
    NodeID string `json:"node_id"`
}

type PodRequest struct {
    CPUs int `json:"cpus"`
    Algorithm string `json:"algorithm"`
}

func main() {
    app := &cli.App{
        Name:  "cluster-cli",
        Usage: "CLI tool for managing the cluster",
        Commands: []*cli.Command{
            {
                Name:  "nodes",
                Usage: "List all nodes in the cluster",
                Action: func(c *cli.Context) error {
                    resp, err := http.Get("http://localhost:8080/nodes")
                    if err != nil {
                        return fmt.Errorf("error sending request: %v", err)
                    }
                    defer resp.Body.Close()

                    body, err := ioutil.ReadAll(resp.Body)
                    if err != nil {
                        return fmt.Errorf("error reading response: %v", err)
                    }

                    if resp.StatusCode != http.StatusOK {
                        return fmt.Errorf("server returned error: %s", string(body))
                    }

                    var nodes []Node
                    if err := json.Unmarshal(body, &nodes); err != nil {
                        return fmt.Errorf("error parsing response: %v", err)
                    }

                    fmt.Printf("\n%-40s %-8s %-8s %-10s %-20s\n", "NODE ID", "CPUs", "USED", "STATUS", "PODS")
                    fmt.Println(strings.Repeat("-", 90))

                    for _, node := range nodes {
                        pods := strings.Join(node.Pods, ", ")
                        if pods == "" {
                            pods = "none"
                        }
                        fmt.Printf("%-40s %-8d %-8d %-10s %-20s\n",
                            node.ID, node.CPUs, node.UsedCPUs, node.Status, pods)
                    }
                    fmt.Println()

                    return nil
                },
            },
            {
                Name:  "add-node",
                Usage: "Add a new node to the cluster",
                Flags: []cli.Flag{
                    &cli.IntFlag{
                        Name:     "cpus",
                        Usage:    "Number of CPUs for the node",
                        Required: true,
                    },
                },
                Action: func(c *cli.Context) error {
                    request := NodeRequest{
                        CPUs: c.Int("cpus"),
                    }

                    jsonData, err := json.Marshal(request)
                    if err != nil {
                        return fmt.Errorf("error marshaling request: %v", err)
                    }

                    resp, err := http.Post("http://localhost:8080/add_node", "application/json", bytes.NewBuffer(jsonData))
                    if err != nil {
                        return fmt.Errorf("error sending request: %v", err)
                    }
                    defer resp.Body.Close()

                    body, err := ioutil.ReadAll(resp.Body)
                    if err != nil {
                        return fmt.Errorf("error reading response: %v", err)
                    }

                    if resp.StatusCode != http.StatusOK {
                        return fmt.Errorf("server returned error: %s", string(body))
                    }

                    fmt.Printf("Node added successfully: %s\n", string(body))
                    return nil
                },
            },
            {
                Name:  "delete-node",
                Usage: "Delete a node from the cluster",
                Flags: []cli.Flag{
                    &cli.StringFlag{
                        Name:     "node-id",
                        Usage:    "ID of the node to delete",
                        Required: true,
                    },
                },
                Action: func(c *cli.Context) error {
                    request := DeleteNodeRequest{
                        NodeID: c.String("node-id"),
                    }

                    jsonData, err := json.Marshal(request)
                    if err != nil {
                        return fmt.Errorf("error marshaling request: %v", err)
                    }

                    req, err := http.NewRequest("DELETE", "http://localhost:8080/delete_node", bytes.NewBuffer(jsonData))
                    if err != nil {
                        return fmt.Errorf("error creating request: %v", err)
                    }
                    req.Header.Set("Content-Type", "application/json")

                    client := &http.Client{}
                    resp, err := client.Do(req)
                    if err != nil {
                        return fmt.Errorf("error sending request: %v", err)
                    }
                    defer resp.Body.Close()

                    body, err := ioutil.ReadAll(resp.Body)
                    if err != nil {
                        return fmt.Errorf("error reading response: %v", err)
                    }

                    if resp.StatusCode != http.StatusOK {
                        return fmt.Errorf("server returned error: %s", string(body))
                    }

                    fmt.Printf("Node deleted successfully: %s\n", string(body))
                    return nil
                },
            },
            {
                Name:  "restart-node",
                Usage: "Restart a node in the cluster",
                Flags: []cli.Flag{
                    &cli.StringFlag{
                        Name:     "node-id",
                        Usage:    "ID of the node to restart",
                        Required: true,
                    },
                },
                Action: func(c *cli.Context) error {
                    request := RestartNodeRequest{
                        NodeID: c.String("node-id"),
                    }

                    jsonData, err := json.Marshal(request)
                    if err != nil {
                        return fmt.Errorf("error marshaling request: %v", err)
                    }

                    req, err := http.NewRequest("PUT", "http://localhost:8080/restart_node", bytes.NewBuffer(jsonData))
                    if err != nil {
                        return fmt.Errorf("error creating request: %v", err)
                    }
                    req.Header.Set("Content-Type", "application/json")

                    client := &http.Client{}
                    resp, err := client.Do(req)
                    if err != nil {
                        return fmt.Errorf("error sending request: %v", err)
                    }
                    defer resp.Body.Close()

                    body, err := ioutil.ReadAll(resp.Body)
                    if err != nil {
                        return fmt.Errorf("error reading response: %v", err)
                    }

                    if resp.StatusCode != http.StatusOK {
                        return fmt.Errorf("server returned error: %s", string(body))
                    }

                    fmt.Printf("Node restarted successfully: %s\n", string(body))
                    return nil
                },
            },
            {
                Name:  "add-pod",
                Usage: "Add a new pod to the cluster",
                Flags: []cli.Flag{
                    &cli.IntFlag{
                        Name:     "cpus",
                        Usage:    "Number of CPUs required for the pod",
                        Required: true,
                    },
                    &cli.StringFlag{
                        Name:     "algorithm",
                        Usage:    "Scheduling algorithm (first_fit, best_fit, worst_fit)",
                    },
                },
                Action: func(c *cli.Context) error {
                    algorithm := c.String("algorithm")
                    validAlgorithms := map[string]bool{
                        "first_fit": true,
                        "best_fit":  true,
                        "worst_fit": true,
                    }
                    if algorithm == "" {
                        algorithm = "first_fit" // Default to first_fit if not specified
                    }
                    if !validAlgorithms[algorithm] {
                        return fmt.Errorf("invalid algorithm: %s. Valid options are: first_fit, best_fit, worst_fit", algorithm)
                    }
                    request := PodRequest{
                        CPUs: c.Int("cpus"),
                        Algorithm: c.String("algorithm"),
                    }

                    jsonData, err := json.Marshal(request)
                    if err != nil {
                        return fmt.Errorf("error marshaling request: %v", err)
                    }

                    resp, err := http.Post("http://localhost:8080/add_pod", "application/json", bytes.NewBuffer(jsonData))
                    if err != nil {
                        return fmt.Errorf("error sending request: %v", err)
                    }
                    defer resp.Body.Close()

                    body, err := ioutil.ReadAll(resp.Body)
                    if err != nil {
                        return fmt.Errorf("error reading response: %v", err)
                    }

                    if resp.StatusCode != http.StatusOK {
                        return fmt.Errorf("server returned error: %s", string(body))
                    }

                    fmt.Printf("Pod added successfully: %s\n", string(body))
                    return nil
                },
            },
        },
    }

    if err := app.Run(os.Args); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
