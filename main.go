package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Blazemeter/crane-hook/pkg"
)

func main() {
	statusErr := &pkg.StatusError{}
	pkg.Execute(statusErr)

	err := pkg.Consolidation(statusErr)
	if err != nil {
		fmt.Printf("\n\n[%s][FAIL] %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
		os.Stdout.Sync() // flush output
		os.Exit(1)
	}
	fmt.Printf("\n\n[%s][PASS] All checks passed successfully, Private Location ready to accept Blazemeter Deployments\n", time.Now().Format("2006-01-02 15:04:05"))
	os.Exit(0)
}
