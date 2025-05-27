package pkg

import (
	"context"
	"fmt"
	"os"
	"time"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

func (statusError *StatusError) checkIngress(cs *ClientSet) {
	clientset := cs.clientset
	ingressType := os.Getenv("KUBERNETES_WEB_EXPOSE_TYPE")
	var ingressNs string
	switch ingressType != "" {
	case ingressType == "ISTIO":
		ingressNs = "istio-system"
	case ingressType == "INGRESS":
		ingressNs = "ingress-nginx"
	default:
		fmt.Printf("\n[%s][error] kubernetes_web_expose_type environment variable is not set or has an invalid value", time.Now().Format("2006-01-02 15:04:05"))
		statusError.IngressStatus = fmt.Errorf("kubernetes_web_expose_type environment variable is not set or has an invalid value")
		return // Exit the function if the ingress type is not set or invalid
	}

	listNamespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("\n[%s][error] listing namespaces: %v", time.Now().Format("2006-01-02 15:04:05"), err)
		statusError.IngressStatus = fmt.Errorf("[%s] error listing namespaces: %v", time.Now().Format("2006-01-02 15:04:05"), err)
	} else {

		nsfound := false
		for _, names := range listNamespaces.Items {
			if ingressNs == names.Name {
				fmt.Printf("\n[%s][INFO] Namespace %s exists for the web service type %s", time.Now().Format("2006-01-02 15:04:05"), names.Name, ingressType)
				nsfound = true
				break // Exit the loop as we found the matching namespace
			}
			continue
		}

		if !nsfound {
			fmt.Printf("\n[%s][error] namespace %s does not exist for the web service type %s", time.Now().Format("2006-01-02 15:04:05"), ingressNs, ingressType)
			statusError.IngressStatus = fmt.Errorf("[%s] namespace %s does not exist for the web service type %s", time.Now().Format("2006-01-02 15:04:05"), ingressNs, ingressType)
		}
		if nsfound {
			err := checkIngressResouces(ingressNs, clientset)
			if err != nil {
				fmt.Printf("\n[%s][error] checking ingress resources: %v", time.Now().Format("2006-01-02 15:04:05"), err)
				statusError.IngressStatus = fmt.Errorf("[%s] error checking ingress resources: %v", time.Now().Format("2006-01-02 15:04:05"), err)
			}
			err = checkSecret(clientset)
			if err != nil {
				fmt.Printf("\n[%s][error] checking ingress secret: %v", time.Now().Format("2006-01-02 15:04:05"), err)
				statusError.IngressStatus = fmt.Errorf("[%s] error checking ingress secret: %v", time.Now().Format("2006-01-02 15:04:05"), err)
			}

			if ingressType == "ISTIO" {
				roleErr := ingressRoleCheckIstio(clientset)
				if roleErr != nil {
					fmt.Printf("\n[%s][error] checking istio-system role: %v", time.Now().Format("2006-01-02 15:04:05"), roleErr)
					statusError.IngressStatus = fmt.Errorf("[%s] error checking istio-system role: %v", time.Now().Format("2006-01-02 15:04:05"), roleErr)
				}
				labelErr := labelCheckIstio(clientset)
				if labelErr != nil {
					fmt.Printf("\n[%s][error] checking istio ingress labels: %v", time.Now().Format("2006-01-02 15:04:05"), labelErr)
					statusError.IngressStatus = fmt.Errorf("[%s] error checking istio ingress labels: %v", time.Now().Format("2006-01-02 15:04:05"), labelErr)
				}
				gatewayErr := gatewayCheck()
				if gatewayErr != nil {
					fmt.Printf("\n[%s][error] checking ingress gateway: %v", time.Now().Format("2006-01-02 15:04:05"), gatewayErr)
					statusError.IngressStatus = fmt.Errorf("[%s] error checking ingress gateway: %v", time.Now().Format("2006-01-02 15:04:05"), gatewayErr)
				}
			}
			if ingressType == "INGRESS" {
				err := ingressRoleCheckNginx(clientset)
				if err != nil {
					fmt.Printf("\n[%s][error] checking nginx-ingress role: %v", time.Now().Format("2006-01-02 15:04:05"), err)
					statusError.IngressStatus = fmt.Errorf("[%s] error checking nginx-ingress role: %v", time.Now().Format("2006-01-02 15:04:05"), err)
				}
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
			fmt.Printf("\n[%s][INFO] Service %s is of type LoadBalancer", time.Now().Format("2006-01-02 15:04:05"), svcType.Name)
			lbfound = true
			// Check if the service has an external IP
			if len(svcType.Status.LoadBalancer.Ingress) > 0 {
				fmt.Printf("\n[%s][INFO] Service %s has an external IP: %s", time.Now().Format("2006-01-02 15:04:05"), svcType.Name, svcType.Status.LoadBalancer.Ingress[0].IP)
				externalIpfound = true
			}
		}
	}

	if !lbfound {
		return fmt.Errorf("\n[%s] loadbalancer service type not found in the namespace %s", time.Now().Format("2006-01-02 15:04:05"), ingressNs)
	}
	if !externalIpfound {
		return fmt.Errorf("\n[%s] external ip not found for the loadbalancer service type in the namespace %s", time.Now().Format("2006-01-02 15:04:05"), ingressNs)
	}
	return nil
}

func ingressRoleCheckNginx(clientset *kubernetes.Clientset) error {
	// Read the namespace from the WORKING_NAMESPACE environment variable
	workingNs := os.Getenv("WORKING_NAMESPACE")
	if workingNs == "" {
		return fmt.Errorf("working_namespace environment variable is not set")
	}

	// Read the role name from the ROLE_NAME environment variable
	roleName := os.Getenv("ROLE_NAME")
	if roleName == "" {
		return fmt.Errorf("role_name environment variable is not set")
	}

	// Fetch the specific role by name
	role, err := clientset.RbacV1().Roles(workingNs).Get(context.TODO(), roleName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get role %s in namespace %s: %v", roleName, workingNs, err)
	}

	// Define the required rules
	requiredApiGroup := "networking.k8s.io"
	requiredResources := []string{"virtualservices", "gateways", "ingresses"}
	requiredVerbs := []string{"get", "list", "create", "delete", "patch", "update"}

	// Check the role's rules for the required permissions
	found := false
	for _, rule := range role.Rules {
		if contains(rule.APIGroups, requiredApiGroup) &&
			containsAll(rule.Resources, requiredResources) &&
			containsAll(rule.Verbs, requiredVerbs) {
			fmt.Printf("\n[%s][INFO] Role %s in namespace %s has the required permissions to run SV with Nginx", time.Now().Format("2006-01-02 15:04:05"), roleName, workingNs)
			found = true
			break
		}
	}
	if !found {
		//fmt.Printf("Role %s in namespace %s does NOT have the required permissions to run SV with Nginx. Rules:\n", roleName, workingNs)
		return fmt.Errorf("role %s in namespace %s does not have the required permissions to run sv with nginx", roleName, workingNs)
	}
	return nil

}

func ingressRoleCheckIstio(clientset *kubernetes.Clientset) error {
	// Read the namespace from the WORKING_NAMESPACE environment variable
	workingNs := os.Getenv("WORKING_NAMESPACE")
	if workingNs == "" {
		return fmt.Errorf("working_namespace environment variable is not set")
	}

	// Read the role name from the ROLE_NAME environment variable
	roleName := os.Getenv("ROLE_NAME")
	if roleName == "" {
		return fmt.Errorf("role_name environment variable is not set")
	}

	// Fetch the specific role by name
	role, err := clientset.RbacV1().Roles(workingNs).Get(context.TODO(), roleName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get role %s in namespace %s: %v", roleName, workingNs, err)
	}

	// Define the required rules
	requiredApiGroup := "networking.istio.io"
	requiredResources := []string{"destinationrules", "virtualservices", "gateways"}
	requiredVerbs := []string{"get", "list", "create", "delete", "patch", "update"}

	// Check the role's rules for the required permissions
	found := false
	for _, rule := range role.Rules {
		if contains(rule.APIGroups, requiredApiGroup) &&
			containsAll(rule.Resources, requiredResources) &&
			containsAll(rule.Verbs, requiredVerbs) {
			fmt.Printf("\n[%s][INFO] Role %s in namespace %s has the required permissions to run SV with Istio", time.Now().Format("2006-01-02 15:04:05"), roleName, workingNs)
			found = true
			break
		}
	}
	if !found {
		//fmt.Printf("Role %s in namespace %s does NOT have the required permissions to run SV with Istio. Rules:\n", roleName, workingNs)
		return fmt.Errorf("role %s in namespace %s does not have the required permissions to run sv with istio", roleName, workingNs)
	}
	return nil
}

func gatewayCheck() error {
	workingNs := os.Getenv("WORKING_NAMESPACE")
	// Create Kubernetes REST configuration
	config, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes rest config: %v", err)
	}
	// Create a dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %v", err)
	}

	// Read the gateway name from the environment variable
	gatewayName := os.Getenv("KUBERNETES_ISTIO_GATEWAY_NAME")
	if gatewayName == "" {
		return fmt.Errorf("kubernetes_web_expose_type environment variable is not set")
	}

	// Fetch the Gateway resource from the Istio API group
	gateway, err := dynamicClient.Resource(
		schema.GroupVersionResource{
			Group:    "networking.istio.io",
			Version:  "v1beta1",
			Resource: "gateways",
		},
	).Namespace(workingNs).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get Gateway %s in namespace %s: %v", gatewayName, workingNs, err)
	}

	// Extract the spec from the Gateway resource
	spec, found, err := unstructured.NestedMap(gateway.Object, "spec")
	if err != nil || !found {
		return fmt.Errorf("failed to extract spec from Gateway %s: %v", gatewayName, err)
	}

	// Validate the spec fields
	if err := validateGatewaySpec(spec); err != nil {
		fmt.Printf("gateway %s spec validation failed: %v", gatewayName, err)
		return fmt.Errorf("gateway %s spec validation failed: %v", gatewayName, err)
	}

	fmt.Printf("\n[%s][INFO] Gateway %s in namespace %s matches the required spec", time.Now().Format("2006-01-02 15:04:05"), gatewayName, workingNs)
	return nil
}

func validateGatewaySpec(spec map[string]interface{}) error {
	wildcardCredential := os.Getenv("KUBERNETES_WEB_EXPOSE_TLS_SECRET_NAME")

	selector, ok := spec["selector"].(map[string]interface{})
	if !ok || selector["istio"] != "ingressgateway" {
		return fmt.Errorf("selector.istio must be 'ingressgateway'")
	}

	servers, ok := spec["servers"].([]interface{})
	if !ok || len(servers) < 3 {
		return fmt.Errorf("spec.servers must contain at least 3 entries")
	}

	for _, s := range servers {
		server, ok := s.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid server entry in spec.servers")
		}
		port, ok := server["port"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("port is not a map[string]interface{}")
		}
		number, err := getPortNumber(port["number"])
		if err != nil {
			return err
		}
		protocol, ok := port["protocol"].(string)
		if !ok {
			return fmt.Errorf("protocol is not a string")
		}

		switch number {
		case 80:
			if protocol != "HTTP" {
				return fmt.Errorf("port 80 must use protocol http")
			}
		case 443:
			if protocol != "HTTPS" {
				return fmt.Errorf("port 443 must use protocol https")
			}
			tls, err := getTLS(server)
			if err != nil {
				return fmt.Errorf("port 443: %v", err)
			}
			if tls["mode"] != "SIMPLE" {
				return fmt.Errorf("port 443 must use tls mode simple")
			}
			if tls["credentialName"] != wildcardCredential {
				return fmt.Errorf("port 443 must use the wildcard credential named %s", wildcardCredential)
			}
		case 15443:
			if protocol != "HTTPS" {
				return fmt.Errorf("port 15443 must use protocol https")
			}
			tls, err := getTLS(server)
			if err != nil {
				return fmt.Errorf("port 15443: %v", err)
			}
			if tls["mode"] != "PASSTHROUGH" {
				return fmt.Errorf("port 15443 must use tls mode passthrough")
			}
		}
	}
	return nil
}

// getPortNumber safely extracts the port number as int, otherwise returns an error
func getPortNumber(val interface{}) (int, error) {
	switch n := val.(type) {
	case int:
		return n, nil
	case int64:
		return int(n), nil
	case float64:
		return int(n), nil
	default:
		return 0, fmt.Errorf("port number is not a valid number type")
	}
}

// getTLS safely extracts the tls map from a server entry
func getTLS(server map[string]interface{}) (map[string]interface{}, error) {
	tlsVal, ok := server["tls"]
	if !ok || tlsVal == nil {
		return nil, fmt.Errorf("must have tls configuration")
	}
	tls, ok := tlsVal.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("tls is not a map[string]interface{}")
	}
	return tls, nil
}

func labelCheckIstio(clientset *kubernetes.Clientset) error {
	workingNs := os.Getenv("WORKING_NAMESPACE")
	nsObject, error := clientset.CoreV1().Namespaces().Get(context.TODO(), workingNs, metav1.GetOptions{})
	if error != nil {
		return fmt.Errorf("failed to get namespace %s to check labels: %v", workingNs, error)
	}
	istioInjection := false
	getLabels := nsObject.GetLabels()
	for key, value := range getLabels {
		if key == "istio-injection" && value == "enabled" {
			fmt.Printf("\n[%s][INFO] Namespace %s has the istio-injection label set to enabled", time.Now().Format("2006-01-02 15:04:05"), workingNs)
			istioInjection = true
			return nil
		}
		continue
	}
	if !istioInjection {
		return fmt.Errorf("namespace %s does not have the istio-injection label set to enabled", workingNs)
	}
	return nil
}

func checkSecret(clientset *kubernetes.Clientset) error {
	// Read the namespace from the WORKING_NAMESPACE environment variable
	ingressType := os.Getenv("KUBERNETES_WEB_EXPOSE_TYPE")
	if ingressType == "" {
		return fmt.Errorf("kubernetes_web_expose_type environment variable is not set")
	}
	workingNs := os.Getenv("WORKING_NAMESPACE")
	if workingNs == "" {
		return fmt.Errorf("working_namespace environment variable is not set")
	}

	// Read the secret name from the SECRET_NAME environment variable
	secretName := os.Getenv("KUBERNETES_WEB_EXPOSE_TLS_SECRET_NAME")
	if secretName == "" {
		return fmt.Errorf("kubernetes_web_expose_tls_secret_name environment variable is not set")
	}
	if ingressType == "ISTIO" {
		secret, err := clientset.CoreV1().Secrets("istio-system").Get(context.TODO(), secretName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get secret %s in namespace instio-ingresss: %v", secretName, err)
		}
		fmt.Printf("\n[%s][INFO] Secret %s is found (%s) in namespace istio-system", time.Now().Format("2006-01-02 15:04:05"), secretName, secret.Name)
	}
	if ingressType == "INGRESS" {
		// Fetch the specific secret by name
		secret, err := clientset.CoreV1().Secrets(workingNs).Get(context.TODO(), secretName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get secret %s in namespace %s: %v", secretName, workingNs, err)
		}
		fmt.Printf("\n[%s][INFO] Secret %s is found (%s) in namespace %s", time.Now().Format("2006-01-02 15:04:05"), secretName, secret.Name, workingNs)
	}
	return nil
}
