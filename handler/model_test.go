package handler

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/k0kubun/pp/v3"
)

func TestModel(t *testing.T) {
	root := &Node{
		Name:  "root",
		Range: []int{0, 1},
		Children: map[string]*Node{
			"0": {
				Name:  "node-0",
				Range: []int{0, 1},
				Children: map[string]*Node{
					"0": {
						Name:  "node-0-0",
						Range: []int{1, 2},
						Children: map[string]*Node{
							"1": {
								Name:  "node-0-0-1",
								Range: []int{2, 3},
								Children: map[string]*Node{
									"2": {Name: "leaf-2", Range: []int{0, 7}},
									"3": {Name: "leaf-3", Range: []int{0, 3}},
								},
							},
						},
					},
					"1": {
						Name:  "node-0-1",
						Range: []int{1, 3},
						Children: map[string]*Node{
							"1": {Name: "leaf-1", Range: []int{6, 9}},
							"2": {Name: "leaf-2", Range: []int{2, 63}},
							"3": {Name: "leaf-3", Range: []int{3, 4}},
						},
					},
				},
			},
			"1": {Name: "node-1", Range: []int{0, 100}},
			"2": {Name: "node-2", Range: []int{10, 20}},
		},
	}

	m := NodeToMap(root)
	fmt.Printf("transformed:\n%#v\n\n", m)

	convertedNode, err := MapToNode(m)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Correct?", reflect.DeepEqual(root, convertedNode))
}

func FillRandomValues(node *Node) error {
	rand.Seed(time.Now().UnixNano())
	return fillRandom(node)
}

func fillRandom(n *Node) error {
	if len(n.Children) == 0 {
		if n.Range[0] > n.Range[1] {
			return fmt.Errorf("invalid range: %v", n.Range)
		}
		n.Value = rand.Intn(n.Range[1]-n.Range[0]+1) + n.Range[0]
		return nil
	}

	for _, child := range n.Children {
		if err := fillRandom(child); err != nil {
			return err
		}
	}
	return nil
}
func TestGenerateRandomAction(t *testing.T) {

	for i := 0; i < 1; i++ {
		podNode, err := StructToNode[InjectionConf]()
		if err != nil {
			t.Error(err)
		}
		FillRandomValues(podNode)
		_, err = NodeToStruct[InjectionConf](podNode)
		if err != nil {
			t.Error(err)
		}

		m := NodeToMap(podNode)
		fmt.Printf("transformed:\n%#v\n\n", m)

		mappedNode, err := MapToNode(m)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("Correct?", reflect.DeepEqual(podNode, mappedNode))
	}
}

func TestHumanizeMap(t *testing.T) {
	m := map[string]any{
		"children": map[string]any{
			"0": map[string]any{
				"value": 1,
			},
			"1": map[string]any{
				"value": 0,
			},
			"2": map[string]any{
				"value": 42,
			},
			"3": map[string]any{
				"value": 10,
			},
			"4": map[string]any{
				"value": 5,
			},
			"5": map[string]any{
				"value": 3,
			},
		},
	}

	newMap, err := HumanizeMap(25, m)
	if err != nil {
		t.Error(err.Error())
		return
	}

	pp.Println(newMap)
}

func TestUnhumanizeMap(t *testing.T) {
	m := map[string]any{
		"Duration":  1,
		"Namespace": "ts",
		"AppName":   "ts-ui-dashboard",
		"Spec": map[string]any{
			"LatencyMs":  10,
			"SQLType":    3,
			"TableIndex": 5,
		},
	}

	newMap, err := UnhumanizeMap(25, m)
	if err != nil {
		t.Error(err.Error())
		return
	}

	pp.Println(newMap)
}
