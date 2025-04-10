package node

import (
    "fmt"
    // "time"
    "sort"
    "cluster-sim/internal/pod"
    // "github.com/google/uuid"
    "log"
)

// SchedulePodBestFit implements best-fit scheduling: chooses the node with the smallest leftover capacity.
func SchedulePodBestFit(pod pod.Pod, nodes map[string]Node) (string, error) {
    nodeList := make([]Node, 0, len(nodes))
    for _, n := range nodes {
        nodeList = append(nodeList, n)
    }

    // Sort nodes by creation time in ascending order.
    sort.Slice(nodeList, func(i, j int) bool {
        return nodeList[i].CreatedAt.Before(nodeList[j].CreatedAt)
    })
    var selectedID string
    minLeftover := int(^uint(0) >> 1) // max int
    for _, n := range nodeList {
        available := n.CPUs - n.UsedCPUs
        if available >= pod.CPUs {
            leftover := available - pod.CPUs
            if leftover < minLeftover {
                minLeftover = leftover
                selectedID = n.ID
            }
        }
    }
    if selectedID == "" {
        return "", fmt.Errorf("no available nodes with sufficient resources")
    }
    n := nodes[selectedID]
    n.Pods = append(n.Pods, pod.ID)
    n.UsedCPUs += pod.CPUs
    nodes[selectedID] = n
    return selectedID, nil
}
// SchedulePodFirstFit implements first-fit scheduling: assigns the pod to the first node with sufficient resources.
func SchedulePodFirstFit(pod pod.Pod, nodes map[string]Node) (string, error) {
    // Gather nodes into a slice.
    nodeList := make([]Node, 0, len(nodes))
    for _, n := range nodes {
        nodeList = append(nodeList, n)
    }

    // Sort nodes by creation time in ascending order.
    sort.Slice(nodeList, func(i, j int) bool {
        return nodeList[i].CreatedAt.Before(nodeList[j].CreatedAt)
    })

    // Iterate over the sorted nodes.
    for _, n := range nodeList {
        available := n.CPUs - n.UsedCPUs
        if available >= pod.CPUs {
            // Update node in the nodes map.
            id := n.ID
            updatedNode := nodes[id]
            updatedNode.Pods = append(updatedNode.Pods, pod.ID)
            updatedNode.UsedCPUs += pod.CPUs
            nodes[id] = updatedNode
            return id, nil
        }
    }
    return "", fmt.Errorf("no available nodes with sufficient resources")
}
// SchedulePodWorstFit implements worst-fit scheduling: chooses the node with the largest leftover capacity.
func SchedulePodWorstFit(pod pod.Pod, nodes map[string]Node) (string, error) {
    nodeList := make([]Node, 0, len(nodes))
    for _, n := range nodes {
        nodeList = append(nodeList, n)
    }

    // Sort nodes by creation time in ascending order.
    sort.Slice(nodeList, func(i, j int) bool {
        return nodeList[i].CreatedAt.Before(nodeList[j].CreatedAt)
    })
    var selectedID string
    maxLeftover := -1
    for _, n := range nodeList {
        available := n.CPUs - n.UsedCPUs
        if available >= pod.CPUs {
            leftover := available - pod.CPUs
            if leftover > maxLeftover {
                maxLeftover = leftover
                selectedID = n.ID
            }
        }
    }
    if selectedID == "" {
        return "", fmt.Errorf("no available nodes with sufficient resources")
    }
    n := nodes[selectedID]
    n.Pods = append(n.Pods, pod.ID)
    n.UsedCPUs += pod.CPUs
    nodes[selectedID] = n
    return selectedID, nil
}
func SchedulePod(pod pod.Pod, nodes map[string]Node, algorithm string) (string, error) {
    switch algorithm {
    case "best_fit":
        return SchedulePodBestFit(pod, nodes)
    case "worst_fit":
        return SchedulePodWorstFit(pod, nodes)
    case "first_fit":
        fallthrough
    default:
        return SchedulePodFirstFit(pod, nodes)
    }
}

func (nm *NodeManager) reschedulePods(failedNodeID string) {
    nm.Mu.Lock()
    var podsToReschedule []string
    if nodeObj, exists := nm.Nodes[failedNodeID]; exists {
        podsToReschedule = nodeObj.Pods
    } else {
        // If the node is gone, assume we recorded its pods before deletion.
        // Here, you might store such info elsewhere. For simplicity, we'll scan all pods.
        for _, p := range nm.Pods {
            if p.NodeID == failedNodeID {
                podsToReschedule = append(podsToReschedule, p.ID)
            }
        }
    }
    nm.Mu.Unlock()

    for _, podID := range podsToReschedule {
        nm.Mu.Lock()
        p, exists := nm.Pods[podID]
        if !exists {
            nm.Mu.Unlock()
            continue
        }
        // Clear current assignment
        p.NodeID = ""
        p.Status = "Pending"
        nm.Pods[podID] = p
        nm.Mu.Unlock()

        nm.Mu.Lock()
        newNodeID, err := SchedulePod(p, nm.Nodes, "first_fit")
        if err == nil {
            p.NodeID = newNodeID
            p.Status = "Running"
            nm.Pods[podID] = p
            // Also update the destination node.
            // nodeUpdate := nm.Nodes[newNodeID]
            // alreadyPresent:=false;
            // for _,id:=range nodeUpdate.Pods{
            //     if id==podID {
            //         alreadyPresent=true;
            //         break;
            //     }
            // }
            // if !alreadyPresent{
            //     nodeUpdate.Pods = append(nodeUpdate.Pods, podID)
            // }
            // nodeUpdate.UsedCPUs += p.CPUs
            // nm.Nodes[newNodeID] = nodeUpdate
            log.Printf("Pod %s rescheduled to node %s", podID, newNodeID)
        } else {
            log.Printf("Failed to reschedule pod %s: %v", podID, err)
        }
        nm.Mu.Unlock()
    }
}
