package main

import (
	"fmt"
	"os"
	"time"

	"github.com/immnan/crane-testhook/pkg"
)

func main() {
	statusErr := &pkg.StatusError{}
	pkg.Execute(statusErr)

	err := pkg.Consolidation(statusErr)
	if err != nil {
		fmt.Printf("\n\n[%s][error] K8s setup does not meet the requirement to run Blazemeter resources. \n%v\n", time.Now().Format("2006-01-02 15:04:05"), err)
		os.Stdout.Sync() // flush output
		os.Exit(1)
	}
	fmt.Printf("\n\n[%s][DONE] All checks passed successfully, Private Location ready to accept Blazemeter Deployments\n", time.Now().Format("2006-01-02 15:04:05"))
	os.Exit(0)
}
