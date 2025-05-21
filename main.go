package main

import (
	"fmt"
	"time"

	"github.com/immnan/crane-testhook/pkg"
)

func main() {
	statusErr := &pkg.StatusError{}
	pkg.Execute(statusErr)

	err := pkg.Consolidation(statusErr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n\n[%s] All checks passed successfully", time.Now())
}
