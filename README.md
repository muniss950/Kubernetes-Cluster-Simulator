# Kubernetes-Cluster-Simulator
--- 
- ### Build the cli
```
  go build -o cluster-cli cmd/cli.go
```
- ### List all nodes
```
  ./cluster-cli nodes
```
- ### Add a new node with 3 CPUs
```
  ./cluster-cli add-node --cpus 3
```
- ### Delete a node
```
  ./cluster-cli delete-node --node-id "node_container_ce27d8ec-5cf7-43ad-80c4-0aabd089d608"
```
- ### Add a new pod with 3 CPUs following any algorithm (best_fit,worst_fit,first_fit;default is first fit)
```
  ./cluster-cli add-pod --cpus 3 --algorithm best_fit
```
- ### Restart a node
```
  ./cluster-cli restart-node --node-id "node_container_9c134f04-f5b3-475b-a6ac-7d53861652b3"
```
