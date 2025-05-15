package main

import (
	"errors"
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {

	statusErr := StatusError{}
	fmt.Println("Starting the requirements check...")
	cs := &clientSet{}
	cs.getClientSet()
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
	statusErr.listNodesDetails(clientSet{})

	ingressType := os.Getenv("KUBERNETES_WEB_EXPOSE_TYPE")
	if ingressType != "" {
		statusErr.checkIngress(clientSet{})
	}

	err := consolidation(&statusErr)
	if err != nil {
		//fmt.Println("\n")
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
	IngressStatus              error
}

type clientSet struct {
	clientset *kubernetes.Clientset
}

func (cs *clientSet) getClientSet() {
	// Create a new Kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	//return clientset
	cs.clientset = clientset
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
	// Check IngressAvailability
	if statusErr.IngressStatus != nil {
		return errors.New("requirements check failed")
	}
	return nil
}

// Helper functions:
// contains checks if a string is present in a slice.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// containsAll checks if all elements of `subset` are present in `set`.
func containsAll(set []string, subset []string) bool {
	for _, sub := range subset {
		found := false
		for _, s := range set {
			if s == sub {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
