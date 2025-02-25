package main

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {

	//statusErr := StatusError{}
	fmt.Println("Starting the requirements check...")
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	//	wg.add(1)
	listNodesDetails(clientset)
	netWorkCheckBlaze()

}

//namespace := "default" // replace with your namespace

type StatusError struct {
	NodeStatus    error
	NetworkStatus []map[string]error
}

//type Status struct {
//	Network   networkCheck
//	SystemReq systemRequirements
//	k8s       k8sSetupRequirements
//}
//
//type networkCheck struct {
//	Blazemeter    []string
//	ImageRegistry bool
//	DockerHub     bool
//	Pip           bool
//	GoogleAPIs    bool
//}
//
//type systemRequirements struct {
//	NodeCount int
//	CPU       string
//	MEM       string
//	Volume    string
//}
//
//type k8sSetupRequirements struct {
//}
