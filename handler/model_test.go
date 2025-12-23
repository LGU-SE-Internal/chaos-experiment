package handler

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/k0kubun/pp/v3"
)

func TestGenerateRandomAction(t *testing.T) {
	rand.Seed(time.Now().UnixNano()) // For reproducibility in tests
	for range 1000 {
		spec, err := StructToNode[InjectionConf](context.Background(), SystemTrainTicket)
		if err != nil {
			t.Error(err)
		}

		podNode, err := randomGenerateNode(spec)
		if err != nil {
			t.Error(err)
		}

		conf, err := NodeToStruct[InjectionConf](context.Background(), podNode)
		if err != nil {
			t.Error(err)
		}

		pp.Println(conf.GetDisplayConfig(context.Background()))
	}
}

func randomGenerateNode(spec *Node) (*Node, error) {
	if spec.Children == nil {
		return nil, fmt.Errorf("specNode must have children defined")
	}

	spec, err := deepCopy(spec)
	if err != nil {
		return nil, fmt.Errorf("error deep copying spec: %w", err)
	}

	keys := make([]string, 0, len(spec.Children))
	for key := range spec.Children {
		keys = append(keys, key)
	}

	if len(keys) == 0 {
		return nil, errors.New("no children keys available")
	}

	chosenKey := keys[rand.Intn(len(keys))]
	childNode := spec.Children[chosenKey]

	if childNode == nil {
		return nil, fmt.Errorf("child node for key %s not found in spec", chosenKey)
	}

	value, err := strconv.Atoi(chosenKey)
	if err != nil {
		return nil, fmt.Errorf("cannot convert key %s to int: %v", chosenKey, err)
	}

	node, err := fillNode(chosenKey, childNode)
	if err != nil {
		return nil, fmt.Errorf("error filling node %s: %w", chosenKey, err)
	}

	res := &Node{
		Children: map[string]*Node{
			chosenKey: node,
		},
		Value: value,
	}

	return res, nil
}

func deepCopy(node *Node) (*Node, error) {
	if node == nil {
		return nil, fmt.Errorf("node is nil")
	}

	copiedNode := &Node{
		Name:        node.Name,
		Description: node.Description,
		Range:       make([]int, len(node.Range)),
		Value:       node.Value,
	}

	copy(copiedNode.Range, node.Range)

	if len(node.Children) > 0 {
		copiedNode.Children = make(map[string]*Node, len(node.Children))
		for key, child := range node.Children {
			copiedChild, err := deepCopy(child)
			if err != nil {
				return nil, fmt.Errorf("error copying child %s: %w", key, err)
			}

			copiedNode.Children[key] = copiedChild
		}
	}

	return copiedNode, nil
}

func fillNode(name string, node *Node) (*Node, error) {
	if node.Children != nil {
		children := make(map[string]*Node, len(node.Children))
		for key, subNode := range node.Children {
			newNode, err := fillNode(key, subNode)
			if err != nil {
				return nil, fmt.Errorf("error filling node %s: %w", key, err)
			}

			children[key] = newNode
		}

		return &Node{
			Children:    children,
			Name:        name,
			Range:       node.Range,
			Description: node.Description,
		}, nil
	} else if node.Range != nil && len(node.Range) == 2 {
		if node.Range[0] > node.Range[1] {
			return nil, fmt.Errorf("node range must be defined with a lower bound less than the upper bound")
		}

		value := rand.Intn(node.Range[1]-node.Range[0]+1) + node.Range[0]
		tempNode := &Node{Value: value}
		if name == "1" {
			tempNode.Value = 0
		}

		return tempNode, nil
	} else {
		return nil, fmt.Errorf("node %s has no children or range defined", name)
	}
}
