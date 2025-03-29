package handler

import (
	"fmt"
	"testing"
)

func TestGenerateActionSpace(t *testing.T) {

	chaosSpec := SpecMap[CPUStress]
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
		"CPULoad":   100,
		"CPUWorker": 2,
		"Duration": 2,
	}
	err = ValidateAction(manualAction, actionSpace)
	if err != nil {
		fmt.Println("Validation Error (Manual):", err)
	} else {
		fmt.Println("Manual Action is valid!")
	}

	chaosSpec, err = ActionToStruct(CPUStress, manualAction)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Converted ChaosSpec: %+v\n", chaosSpec)
	}
}
