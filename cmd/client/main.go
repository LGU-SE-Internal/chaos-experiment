package main

import (
	"github.com/LGU-SE-Internal/chaos-experiment/client"
	"github.com/k0kubun/pp/v3"
)

func main() {
	list, _ := client.GetContainersWithAppLabel("ts")
	pp.Print(list)
}
