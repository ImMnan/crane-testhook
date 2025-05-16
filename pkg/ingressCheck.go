package pkg

import (
	"context"
	"fmt"
	"os"

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
		ingressNs = "contour"
	}

	listNamespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("error listing namespaces: %v", err)
		statusError.IngressStatus = fmt.Errorf("error listing namespaces: %v", err)
	} else {

		nsfound := false
		for _, names := range listNamespaces.Items {
			if ingressNs == names.Name {
				fmt.Printf("Namespace %s exists for the web service type %s\n", names.Name, ingressType)
				nsfound = true
				break // Exit the loop as we found the matching namespace
			}
			continue
		}

		if !nsfound {
			fmt.Printf("namespace %s does not exist for the web service type %s", ingressNs, ingressType)
			statusError.IngressStatus = fmt.Errorf("namespace %s does not exist for the web service type %s", ingressNs, ingressType)

		}
		if nsfound {
			err := checkIngressResouces(ingressNs, clientset)
			if err != nil {
				fmt.Printf("error checking ingress resources:\n %v", err)
				statusError.IngressStatus = err
			}
			if ingressType == "ISTIO" {
				roleErr := ingressRoleCheckIstio(clientset)
				if roleErr != nil {
					fmt.Printf("error checking istio-ingress role:\n %v", roleErr)
					statusError.IngressStatus = err
				}
				labelErr := labelCheckIstio(clientset)
				if labelErr != nil {
					fmt.Printf("error checking istio ingress labels:\n %v", labelErr)
					statusError.IngressStatus = labelErr
				}
				gatewayErr := gatewayCheck(ingressNs)
				if gatewayErr != nil {
					fmt.Printf("error checking ingress gateway:\n %v", gatewayErr)
					statusError.IngressStatus = gatewayErr
				}
			}
			if ingressType == "INGRESS" {
				err := ingressRoleCheckNginx(clientset)
				if err != nil {
					fmt.Printf("error checking nginx-ingress role:\n %v", err)
					statusError.IngressStatus = err
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
	for _, rule := range role.Rules {
		if contains(rule.APIGroups, requiredApiGroup) &&
			containsAll(rule.Resources, requiredResources) &&
			containsAll(rule.Verbs, requiredVerbs) {
			fmt.Printf("Role %s in namespace %s has the required permissions to run SV with Nginx\n", roleName, workingNs)
			return nil // Found a matching rule
		} else {
			return fmt.Errorf("role %s in namespace %s does not have the required permissions to run sv with nginx", roleName, workingNs)
		}
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
	for _, rule := range role.Rules {
		if contains(rule.APIGroups, requiredApiGroup) &&
			containsAll(rule.Resources, requiredResources) &&
			containsAll(rule.Verbs, requiredVerbs) {
			fmt.Printf("Role %s in namespace %s has the required permissions to run SV with Istio\n", roleName, workingNs)
			return nil // Found a matching rule
		} else {
			return fmt.Errorf("role %s in namespace %s does not have the required permissions to run sv with istio", roleName, workingNs)
		}
	}
	return nil
}

func gatewayCheck(ingressNs string) error {

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
	gatewayName := os.Getenv("KUBERNETES_WEB_EXPOSE_TYPE")
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
	).Namespace(ingressNs).Get(context.TODO(), gatewayName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get Gateway %s in namespace %s: %v", gatewayName, ingressNs, err)
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

	fmt.Printf("Gateway %s in namespace %s matches the required spec\n", gatewayName, ingressNs)
	return nil
}

// validateGatewaySpec validates the Gateway spec against the required structure
func validateGatewaySpec(spec map[string]interface{}) error {

	wildcardCredential := os.Getenv("KUBERNETES_WEB_EXPOSE_TLS_SECRET_NAME")
	// Check the selector
	selector, found := spec["selector"].(map[string]interface{})
	if !found || selector["istio"] != "ingressgateway" {
		return fmt.Errorf("selector.istio must be 'ingressgateway'")
	}

	// Check the servers
	servers, found := spec["servers"].([]interface{})
	if !found || len(servers) < 3 {
		return fmt.Errorf("spec.servers must contain at least 3 entries")
	}

	// Validate each server
	for _, server := range servers {
		serverMap, ok := server.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid server entry in spec.servers")
		}

		port := serverMap["port"].(map[string]interface{})
		number := int(port["number"].(float64))
		protocol := port["protocol"].(string)
		tls := serverMap["tls"].(map[string]interface{})

		switch number {
		case 80:
			if protocol != "HTTP" {
				return fmt.Errorf("port 80 must use protocol http")
			}
		case 443:
			if protocol != "HTTPS" && serverMap["tls"].(map[string]interface{})["mode"] != "SIMPLE" {
				//	fmt.Errorf("port 443 must use protocol HTTPS with TLS mode SIMPLE")
				return fmt.Errorf("port 443 must use protocol https with tls mode simple")
			}
			if tls["credentialName"] != wildcardCredential {
				//fmt.Errorf("port 443 must use the wildcard credential named %s", wildcardCredential)
				return fmt.Errorf("port 443 must use the wildcard credential named %s", wildcardCredential)
			}
		case 15443:
			if protocol != "HTTPS" && serverMap["tls"].(map[string]interface{})["mode"] != "PASSTHROUGH" {
				//fmt.Errorf("port 15443 must use protocol HTTPS with TLS mode PASSTHROUGH")
				return fmt.Errorf("port 15443 must use protocol https with tls mode passthrough")
			}
		default:
			continue
		}
	}

	return nil
}

func labelCheckIstio(clientset *kubernetes.Clientset) error {
	ns := os.Getenv("WORKING_NAMESPACE")
	nsObject, error := clientset.CoreV1().Namespaces().Get(context.TODO(), ns, metav1.GetOptions{})
	if error != nil {
		return fmt.Errorf("failed to get namespace %s to check labels: %v", ns, error)
	}
	istioInjection := false
	getLabels := nsObject.GetLabels()
	for key, value := range getLabels {
		if key == "istio-injection" && value == "enabled" {
			fmt.Printf("Namespace %s has the istio-injection label set to enabled\n", ns)
			istioInjection = true
			return nil
		}
		continue
	}
	if !istioInjection {
		return fmt.Errorf("namespace %s does not have the istio-injection label set to enabled", ns)
	}
	return nil
}
