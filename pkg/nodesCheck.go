package pkg

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (statusErr *StatusError) listNodesDetails(cs *ClientSet) {
	clientset := cs.clientset
	// Configure slog logger
	//statusErr := &StatusError{}
	//logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Infinite loop to get node details every 30 seconds, until the pod/program is terminated.
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		statusErr.NodeStatus = err
		//fmt.Println(statusErr.NodeStatus)
	}
	for i, nd := range nodes.Items {
		if nd.Status.Capacity.Cpu().MilliValue() <= 2000 {
			statusErr.NodeResourceStatus = append(statusErr.NodeResourceStatus, map[string]error{fmt.Sprintf("cpu node %d", i): fmt.Errorf("insufficient %d", nd.Status.Capacity.Cpu().MilliValue())})
			fmt.Printf("node %d insufficient cpu %d\n", i, nd.Status.Capacity.Cpu().MilliValue())
		}
		fmt.Printf("Node: %d, CPU: %d", i, nd.Status.Capacity.Cpu().MilliValue())

		if nd.Status.Capacity.Memory().MilliValue() <= 4096 {
			statusErr.NodeResourceStatus = append(statusErr.NodeResourceStatus, map[string]error{"memory/node": fmt.Errorf("insufficient %d", nd.Status.Capacity.Memory().MilliValue())})
			fmt.Printf("node %d insufficient mem %d\n", i, nd.Status.Capacity.Memory().MilliValue())
		}
		fmt.Printf("Node: %d, MEM: %d", i, nd.Status.Capacity.Memory().MilliValue())
		if nd.Status.Capacity.StorageEphemeral().Value() <= 64000 {
			statusErr.NodeResourceStatus = append(statusErr.NodeResourceStatus, map[string]error{"storage/node": fmt.Errorf("insufficient %d", nd.Status.Capacity.StorageEphemeral().Value())})
			fmt.Printf("node %d insufficient storage %d", i, nd.Status.Capacity.StorageEphemeral().Value())
		}
		fmt.Printf("Node: %d, Storage: %d\n", i, nd.Status.Capacity.StorageEphemeral().Value())

	}
}
