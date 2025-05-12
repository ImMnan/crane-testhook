package main

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (statusError *StatusError) checkIngress() {
	clientset := getClientSet()
	ingressType := os.Getenv("KUBERNETES_WEB_EXPOSE_TYPE")
	var ingressNs string
	switch ingressType != "" {
	case ingressType == "ISTIO":
		ingressNs = "istio-system"
	case ingressType == "INGRESS":
		ingressNs = "ingress-nginx"
	default:
		ingressNs = "contour"
	}

	listNamespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("error listing namespaces: %v", err)
		statusError.IngressAvailability = fmt.Errorf("error listing namespaces: %v", err)
	} else {

		nsfound := false
		for _, names := range listNamespaces.Items {
			if ingressNs == names.Name {
				fmt.Printf("Namespace %s exists for the web service type %s\n", names.Name, ingressType)
				nsfound = true
				break // Exit the loop as we found the matching namespace
			}
		}

		if !nsfound {
			fmt.Printf("namespace %s does not exist for the web service type %s", ingressNs, ingressType)
			statusError.IngressAvailability = fmt.Errorf("namespace %s does not exist for the web service type %s", ingressNs, ingressType)

		}
		if nsfound {
			err := checkIngressResouces(ingressNs, clientset)
			if err != nil {
				statusError.IngressAvailability = err
			}
		}
	}
}

func checkIngressResouces(ingressNs string, clientset *kubernetes.Clientset) error {
	svcList, err := clientset.CoreV1().Services(ingressNs).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	lbfound := false
	externalIpfound := false
	for _, svcType := range svcList.Items {
		if svcType.Spec.Type == "LoadBalancer" {
			fmt.Printf("Service %s is of type LoadBalancer", svcType.Name)
			lbfound = true
			// Check if the service has an external IP
			if len(svcType.Status.LoadBalancer.Ingress) > 0 {
				fmt.Printf("Service %s has an external IP: %s", svcType.Name, svcType.Status.LoadBalancer.Ingress[0].IP)
				externalIpfound = true
			}
		}
	}

	if !lbfound {
		return fmt.Errorf("loadbalancer service type not found in the namespace %s", ingressNs)
	}
	if !externalIpfound {
		return fmt.Errorf("external ip not found for the loadbalancer service type in the namespace %s", ingressNs)
	}
	return nil
}
