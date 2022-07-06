package main

import (
	"fmt"
	"os"

	dics "app/configs"

	"github.com/sarulabs/dingo/v4"
)

func main() {
	err := dingo.GenerateContainerWithCustomPkgName((*dics.Engine)(nil), "generated", "engine")
	if err != nil {
		fmt.Println("Error dumping container: ", err.Error())
		os.Exit(1)
	}
}
