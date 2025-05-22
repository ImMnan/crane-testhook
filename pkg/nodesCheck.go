package pkg

import (
	"context"
	"fmt"
	"time"

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
		var errs []string

		cpu := nd.Status.Capacity.Cpu().MilliValue()
		mem := nd.Status.Capacity.Memory().Value()
		storage := nd.Status.Capacity.StorageEphemeral().Value()
		memMB := mem / (1024 * 1024)
		storageMB := storage / (1024 * 1024)
		if cpu < 2000 {
			statusErr.NodeResourceStatus = append(statusErr.NodeResourceStatus, map[string]error{
				fmt.Sprintf("cpu node %d", i): fmt.Errorf("insufficient %d", cpu),
			})
			errs = append(errs, fmt.Sprintf("CPU: %d", cpu))
		}
		if memMB < 4096 {
			statusErr.NodeResourceStatus = append(statusErr.NodeResourceStatus, map[string]error{
				fmt.Sprintf("memory node %d", i): fmt.Errorf("insufficient %d", memMB),
			})
			errs = append(errs, fmt.Sprintf("MEM: %d MB", memMB))
		}
		if storageMB < 64 {
			statusErr.NodeResourceStatus = append(statusErr.NodeResourceStatus, map[string]error{
				fmt.Sprintf("storage node %d", i): fmt.Errorf("insufficient %d", storageMB),
			})
			errs = append(errs, fmt.Sprintf("STORAGE: %d MB", storageMB))
		}

		if len(errs) > 0 {
			fmt.Printf("\n[%s][error] node %d insufficient resources: %s", time.Now().Format("2006-01-02 15:04:05"), i+1, errs)
		} else {
			fmt.Printf("\n[%s][INFO] Node: %d, CPU: %d, MEM: %d MB, Storage: %d MB", time.Now().Format("2006-01-02 15:04:05"), i+1, cpu, memMB, storageMB)
		}
	}
}
