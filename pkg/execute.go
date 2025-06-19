package pkg

import (
	"errors"
	"fmt"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

//namespace := "default" // replace with your namespace

type StatusError struct {
	NodeStatus                 error
	NodeResourceStatus         []map[string]error
	BlazeNetworkStatus         []map[string]error
	ImageRegistryNetworkStatus error
	ThirdPartyNetworkStatus    []map[string]error
	IngressStatus              error
	RBAC                       error
}

type ClientSet struct {
	clientset *kubernetes.Clientset
}

func (cs *ClientSet) getClientSet() {
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

func Execute(statusError *StatusError) {
	// This is a test function to check if the package is working
	cs := &ClientSet{}
	cs.getClientSet()
	//	statusError := &StatusError{}
	fmt.Printf("\n[%s][INFO] Starting the requirements check...", time.Now().Format("2006-01-02 15:04:05"))
	svEnabled := os.Getenv("KUBERNETES_WEB_EXPOSE_TYPE")
	statusError.networkCheckBlaze()
	statusError.networkCheckImageRegistry()
	statusError.networkCheckThirdParty()
	statusError.listNodesDetails(cs)
	statusError.rbacDefault(cs)
	if svEnabled != "" {
		statusError.checkIngress(cs)
	}

}

func Consolidation(statusErr *StatusError) error {
	// Check NodeStatus
	if statusErr.NodeStatus != nil {
		return errors.New("requirements check failed, check errors in logs")
	}

	// Check NodeResourceStatus
	for _, resourceStatus := range statusErr.NodeResourceStatus {
		for _, err := range resourceStatus {
			if err != nil {
				return errors.New("requirements check failed, check errors in logs")
			}
		}
	}

	// Check BlazeNetworkStatus
	for _, blazeStatus := range statusErr.BlazeNetworkStatus {
		for _, err := range blazeStatus {
			if err != nil {
				return errors.New("requirements check failed, check errors in logs")
			}
		}
	}

	// Check ImageRegistryNetworkStatus
	if statusErr.ImageRegistryNetworkStatus != nil {
		return errors.New("requirements check failed, check errors in logs")
	}

	// Check ThirdPartyNetworkStatus
	for _, thirdPartyStatus := range statusErr.ThirdPartyNetworkStatus {
		for _, err := range thirdPartyStatus {
			if err != nil {
				return errors.New("requirements check failed, check errors in logs")
			}
		}
	}
	// Check IngressAvailability
	if statusErr.IngressStatus != nil {
		return errors.New("requirements check failed, check errors in logs")
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
			//fmt.Printf("\n[%s]Missing required item: %q", sub)
			return false
		}
	}
	return true
}

//func dedup(slice []string) []string {
//	seen := make(map[string]struct{})
//	var result []string
//	for _, s := range slice {
//		if _, ok := seen[s]; !ok {
//			seen[s] = struct{}{}
//			result = append(result, s)
//		}
//	}
//	return result
//}

func dedup(items []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}
