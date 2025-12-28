package main

import (
	"context"

	"github.com/LGU-SE-Internal/chaos-experiment/client"
	"github.com/k0kubun/pp/v3"
)

func main() {

	list, _ := client.GetContainersWithAppLabel(context.Background(), "ts0")
	pp.Print(list)
}
