package client

import (
	"fmt"
	"testing"
)

func TestGetLabel(t *testing.T) {
	labels, err := GetLabels("ts", "app")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(labels)
}
