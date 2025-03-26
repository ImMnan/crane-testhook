package main

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (statusError *StatusError) checkIngress() {
	clientset := getClientSet()

	ingressType := os.Getenv("KUBERNETES_WEB_EXPOSE_TYPE")

	fmt.Println(ingressType)

	listNamespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, names := range listNamespaces.Items {
		if ingressType == names.Name {
			fmt.Printf("Namespace %s exists for the web service type %s", names.Name, ingressType)
		} else {
			statusError.IngressAvailability = fmt.Errorf("namespace %s does not exist for the web service type %s", names.Name, ingressType)
		}
		//fmt.Println(listNamespaces)
	}
}
