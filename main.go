package main

import (
	"errors"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {

	statusErr := StatusError{}
	fmt.Println("Starting the requirements check...\n")
	//config, err := rest.InClusterConfig()
	//if err != nil {
	//	panic(err.Error())
	//}
	//clientset, err := kubernetes.NewForConfig(config)
	//if err != nil {
	//	panic(err.Error())
	//}
	//	wg.add(1)
	//listNodesDetails(clientset)
	statusErr.networkCheckBlaze()
	statusErr.networkCheckImageRegistry()
	statusErr.networkCheckThirdParty()
	statusErr.listNodesDetails()
	statusErr.checkIngress()
	err := consolidation(&statusErr)
	if err != nil {
		fmt.Println("\n")
		panic(err)
	}
	fmt.Println("\nAll checks passed successfully")
}

//namespace := "default" // replace with your namespace

type StatusError struct {
	NodeStatus                 error
	NodeResourceStatus         []map[string]error
	BlazeNetworkStatus         []map[string]error
	ImageRegistryNetworkStatus error
	ThirdPartyNetworkStatus    []map[string]error
	IngressAvailability        error
}

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

func consolidation(statusErr *StatusError) error {
	// Check NodeStatus
	if statusErr.NodeStatus != nil {
		return errors.New("requirements check failed")
	}

	// Check NodeResourceStatus
	for _, resourceStatus := range statusErr.NodeResourceStatus {
		for _, err := range resourceStatus {
			if err != nil {
				return errors.New("requirements check failed")
			}
		}
	}

	// Check BlazeNetworkStatus
	for _, blazeStatus := range statusErr.BlazeNetworkStatus {
		for _, err := range blazeStatus {
			if err != nil {
				return errors.New("requirements check failed")
			}
		}
	}

	// Check ImageRegistryNetworkStatus
	if statusErr.ImageRegistryNetworkStatus != nil {
		return errors.New("requirements check failed")
	}

	// Check ThirdPartyNetworkStatus
	for _, thirdPartyStatus := range statusErr.ThirdPartyNetworkStatus {
		for _, err := range thirdPartyStatus {
			if err != nil {
				return errors.New("requirements check failed")
			}
		}
	}
	return nil
}
