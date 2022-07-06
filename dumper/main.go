package main

import (
	"fmt"
	"os"

	dics "github.com/bimalabs/skeleton/v4/configs"
	"github.com/sarulabs/dingo/v4"
)

func main() {
	err := dingo.GenerateContainerWithCustomPkgName((*dics.Engine)(nil), "generated", "app")
	if err != nil {
		fmt.Println("Error dumping container: ", err.Error())
		os.Exit(1)
	}
}
