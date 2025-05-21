package pkg

import (
	"context"
	"fmt"
	"os"
	"time"

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
	requiredResources := []string{"pods", "services", "daemonsets", "replicasets", "deployments", "deployments/scale", "jobs", "pods/*"}
	requiredVerbs := []string{"get", "list", "create", "delete", "patch", "update", "watch", "deletecollection", "createcollection"}

	roleDesc, err := cs.clientset.RbacV1().Roles(workingNs).Get(context.TODO(), roleName, metav1.GetOptions{})
	if err != nil {
		statusError.RBAC = fmt.Errorf("failed to get role %s in namespace %s: %v", roleName, workingNs, err)
	}
	found := false
	var allApiGroups, allResources, allVerbs []string
	for _, rule := range roleDesc.Rules {
		allApiGroups = append(allApiGroups, rule.APIGroups...)
		allResources = append(allResources, rule.Resources...)
		allVerbs = append(allVerbs, rule.Verbs...)
	}
	allApiGroups = dedup(allApiGroups)
	allResources = dedup(allResources)
	allVerbs = dedup(allVerbs)

	fmt.Printf("Role %s in namespace %s has API Groups: %q\n", roleName, workingNs, allApiGroups)
	fmt.Printf("Role %s in namespace %s has Resources: %v\n", roleName, workingNs, allResources)
	fmt.Printf("Role %s in namespace %s has Verbs: %v\n", roleName, workingNs, allVerbs)

	if containsAll(allApiGroups, requiredApiGroup) &&
		containsAll(allResources, requiredResources) &&
		containsAll(allVerbs, requiredVerbs) {
		found = true
	}

	if found {
		fmt.Printf("\n[%s] Role %s in namespace %s has the required permissions\n", time.Now().Format("2006-01-02 15:04:05"), roleName, workingNs)
	} else {
		fmt.Printf("\n[%s] Role %s in namespace %s does NOT have the required permissions\n", time.Now().Format("2006-01-02 15:04:05"), roleName, workingNs)
		statusError.RBAC = fmt.Errorf("role %s in namespace %s does not have the required permissions", roleName, workingNs)
	}

	rbacBindingError := saBinding(cs)
	if rbacBindingError != nil {
		fmt.Printf("\n[%s] role binding error: %v\n", time.Now().Format("2006-01-02 15:04:05"), rbacBindingError)
		statusError.RBAC = fmt.Errorf("role binding error: %v", rbacBindingError)
	}
}

func saBinding(cs *ClientSet) error {
	workingNs := os.Getenv("WORKING_NAMESPACE")
	if workingNs == "" {
		return fmt.Errorf("WORKING_NAMESPACE environment variable is not set")
	}
	// Read the role name from the ROLE_NAME environment variable
	roleName := os.Getenv("ROLE_NAME")
	if roleName == "" {
		return fmt.Errorf(" environment variable is not set")
	}

	saName := os.Getenv("SERVICE_ACCOUNT_NAME")
	if saName == "" {
		return fmt.Errorf("service_account_name environment variable is not set")
	}

	roleBinding := os.Getenv("ROLE_BINDING_NAME")
	if roleBinding == "" {
		return fmt.Errorf("role_binding_name environment variable is not set")
	}
	// check if the role binding has the correct role and service account
	rbDesc, err := cs.clientset.RbacV1().RoleBindings(workingNs).Get(context.TODO(), roleBinding, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get role binding %s in namespace %s: %v", roleBinding, workingNs, err)
	}
	if rbDesc.RoleRef.Name != roleName {
		return fmt.Errorf("role binding %s does not reference role %s", roleBinding, roleName)
	}
	rolebindingFound := false
	for _, subject := range rbDesc.Subjects {
		if subject.Kind == "ServiceAccount" && subject.Name == saName {
			fmt.Printf("\n[%s] role binding %s in namespace %s binds service account %s with role %s\n", time.Now().Format("2006-01-02 15:04:05"), roleBinding, workingNs, saName, roleName)
			rolebindingFound = true
			return nil
		}
	}
	if !rolebindingFound {
		return fmt.Errorf("role binding %s does not bind service account %s with role %s", roleBinding, saName, roleName)
	}
	return nil
}
