package controllers

import (
	"fmt"
	"testing"
)

func TestGenerateActionSpace(t *testing.T) {
	type ChaosSpec struct {
		CPULoad    int `range:"1-100"`
		MemorySize int `range:"1-262144"`
		Worker     int `range:"1-8192"`
		Target     int `range:"1-2"`
	}

	chaosSpec := ChaosSpec{}
	actionSpace, err := GenerateActionSpace(chaosSpec)
	if err != nil {
		fmt.Println("Error generating action space:", err)
		return
	}
	fmt.Println("Generated Action Space:", actionSpace)

	randomAction := generateRandomAction(actionSpace)
	fmt.Println("Random Action:", randomAction)

	err = ValidateAction(randomAction, actionSpace)
	if err != nil {
		fmt.Println("Validation Error:", err)
	} else {
		fmt.Println("Action is valid!")
	}

	manualAction := map[string]int{
		"CPULoad":    50,
		"MemorySize": 3123,
		"Worker":     1000,
		"Target":     2,
	}
	err = ValidateAction(manualAction, actionSpace)
	if err != nil {
		fmt.Println("Validation Error (Manual):", err)
	} else {
		fmt.Println("Manual Action is valid!")
	}

	newChaosSpec := &ChaosSpec{}

	err = ActionToStruct(manualAction, newChaosSpec)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Converted ChaosSpec: %+v\n", newChaosSpec)
	}
}
