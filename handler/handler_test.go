package handler

import (
	"fmt"
	"math/rand"
	"testing"

	cli "github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/k0kubun/pp"
)

func TestHandler(t *testing.T) {
	node, err := StructToNode[InjectionConf]()
	if err != nil {
		t.Errorf("StructToNode failed: %v", err)
		return
	}
	// Test the node structure
	if node == nil {
		t.Errorf("Expected non-nil node, got nil")
		return
	}

	mapStru := NodeToMap(node)
	if mapStru == nil {
		t.Errorf("Expected non-nil map, got nil")
		return
	}
	pp.Println(mapStru)

	gend, err := genValue(mapStru)
	if err != nil {
		t.Errorf("genValue failed: %v", err)
		return
	}
	newNode, err := MapToNode(gend)
	if err != nil {
		t.Errorf("MapToNode failed: %v", err)
		return
	}
	pp.Println(newNode)

}

func TestHandler2(t *testing.T) {
	chilren := map[int]any{
		0: map[string]any{
			"value": 1,
		},
		1: map[string]any{
			"value": 0,
		},
		2: map[string]any{
			"value": 3,
		},
	}

	node, err := MapToNode(map[string]any{
		"children": map[int]any{
			0: map[string]any{
				"children": chilren,
			},
		},
	})
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	pp.Println(node)

	conf, err := NodeToStruct[InjectionConf](node)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	pp.Println(conf)

	pp.Println(conf.Create(cli.NewK8sClient()))
}

func genValue(m map[string]any) (map[string]any, error) {
	var rangeI []int
	if r, exist := m["range"]; exist {
		var ok bool
		rangeI, ok = r.([]int)
		if !ok {
			return nil, fmt.Errorf("range is not a string")
		}
	} else {
		return nil, fmt.Errorf("range not exist")
	}

	if value, exist := m["children"]; exist && value != nil {
		subMap, ok := value.(map[int]interface{})
		if !ok {
			return nil, fmt.Errorf("value is not a map[string]interface{}")
		}
		for i := rangeI[0]; i <= rangeI[1]; i++ {
			if va, ok := subMap[i].(map[string]interface{}); ok {
				gened, err := genValue(va)
				if err != nil {
					return nil, err
				}
				subMap[i] = gened
			}

		}
		m["children"] = subMap
	}
	m["value"] = rand.Intn(rangeI[1]-rangeI[0]+1) + rangeI[0]

	return m, nil
}
