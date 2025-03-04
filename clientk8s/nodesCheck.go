package clientk8s

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func getClientSet() *kubernetes.Clientset {
	// Create a new Kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}
func listNodesDetails() {
	clientset := getClientSet()
	// Configure slog logger
	statusErr := &StatusError{}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Infinite loop to get node details every 30 seconds, until the pod/program is terminated.
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		statusErr.NodeStatus = err
		//fmt.Println(statusErr.NodeStatus)
	}
	for i, nd := range nodes.Items {

		capacity := map[string]string{
			"cpu":     nd.Status.Capacity.Cpu().String(),
			"memory":  nd.Status.Capacity.Memory().String(),
			"storage": nd.Status.Capacity.StorageEphemeral().String(),
			"pods":    nd.Status.Capacity.Pods().String(),
		}
		if nd.Status.Capacity.Cpu().MilliValue() <= 2000 {
			statusErr.NodeResourceStatus = append(statusErr.NodeResourceStatus, map[string]error{fmt.Sprintf("cpu node %d", i): fmt.Errorf("insufficient %d", nd.Status.Capacity.Cpu().MilliValue())})
		}
		if nd.Status.Capacity.Memory().MilliValue() <= 4096 {
			statusErr.NodeResourceStatus = append(statusErr.NodeResourceStatus, map[string]error{"memory/node": fmt.Errorf("insufficient %d", nd.Status.Capacity.Memory().MilliValue())})
		}
		if nd.Status.Capacity.StorageEphemeral().Value() <= 64000 {
			statusErr.NodeResourceStatus = append(statusErr.NodeResourceStatus, map[string]error{"storage/node": fmt.Errorf("insufficient %d", nd.Status.Capacity.StorageEphemeral().Value())})
		}

		allocatable := map[string]string{
			"cpu":     nd.Status.Allocatable.Cpu().String(),
			"memory":  nd.Status.Allocatable.Memory().String(),
			"storage": nd.Status.Allocatable.StorageEphemeral().String(),
			"pods":    nd.Status.Allocatable.Pods().String(),
		}
		logger.Info("details",
			slog.String("time", time.Now().Format(time.RFC3339)),
			slog.Int("node", i+1),
			slog.String("name", nd.Name),
			slog.Any("allocatable", allocatable),
			slog.Any("capacity", capacity),
		)
	}
}
