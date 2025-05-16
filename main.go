package main

import (
	"fmt"

	"github.com/immnan/crane-testhook/pkg"
)

func main() {
	statusErr := pkg.StatusError{}
	pkg.Execute()

	err := pkg.Consolidation(&statusErr)
	if err != nil {
		panic(err)
	}
	fmt.Println("\nAll checks passed successfully")
}
