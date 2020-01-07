package main

import (
	"github.com/go-chassis/go-chassis"

	_ "github.com/go-chassis/go-chassis-cloud/provider/huawei/engine"
)

func main() {
	err := chassis.Init()
	if err != nil {
		panic(err)
	}
	err = chassis.Run()
	if err != nil {
		panic(err)
	}
}
