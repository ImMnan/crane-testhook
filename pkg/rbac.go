package pkg

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (statusError *StatusError) rbacDefault(cs *ClientSet) {

	workingNs := os.Getenv("WORKING_NAMESPACE")
	if workingNs == "" {
		statusError.RBAC = fmt.Errorf("WORKING_NAMESPACE environment variable is not set")
	}
	// Read the role name from the ROLE_NAME environment variable
	roleName := os.Getenv("ROLE_NAME")
	if roleName == "" {
		statusError.RBAC = fmt.Errorf("ROLE_NAME environment variable is not set")
	}

	requiredApiGroup := []string{"", "extensions", "apps", "batch"}
	requiredResources := []string{"pods", "services", "daemonsets", "replicasets", "deployments", "deployment/scale", "jobs", "pods/*"}
	requiredVerbs := []string{"get", "list", "create", "delete", "patch", "update", "watch", "deletecollection", "createcollection"}

	role, err := cs.clientset.RbacV1().Roles(workingNs).Get(context.TODO(), roleName, metav1.GetOptions{})
	if err != nil {
		statusError.RBAC = fmt.Errorf("failed to get role %s in namespace %s: %v", roleName, workingNs, err)
	}

	for _, rule := range role.Rules {
		if containsAll(rule.APIGroups, requiredApiGroup) &&
			containsAll(rule.Resources, requiredResources) &&
			containsAll(rule.Verbs, requiredVerbs) {
			fmt.Printf("Role %s in namespace %s has the required permissions\n", roleName, workingNs)
			statusError.RBAC = nil // Found a matching rule
		} else {
			fmt.Printf("Role %s in namespace %s does not have the required permissions\n", roleName, workingNs)
			statusError.RBAC = fmt.Errorf("role %s in namespace %s does not have the required permissions", roleName, workingNs)
		}
	}
}
