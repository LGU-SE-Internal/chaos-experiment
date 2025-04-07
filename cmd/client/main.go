package main

import (
	"github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/k0kubun/pp/v3"
)

func main() {
	list,_ := client.GetContainersWithAppLabel("ts")
	pp.Print(list)
}
