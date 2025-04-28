package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/k0kubun/pp"
)

// 测试获取配置
func TestHandler(t *testing.T) {
	if err := InitTargetConfig(map[string]int{"ts": 5}, "app"); err != nil {
		t.Error(err.Error())
		return
	}

	node, err := StructToNode[InjectionConf]("ts")
	if err != nil {
		t.Errorf("StructToNode failed: %v", err)
		return
	}

	// Test the node structure
	if node == nil {
		t.Errorf("Expected non-nil node, got nil")
		return
	}

	mapStru := NodeToMap(node, true)
	if mapStru == nil {
		t.Errorf("Expected non-nil map, got nil")
		return
	}
	pp.Println(mapStru)
}

// 测试创建
func TestHandler2(t *testing.T) {
	targetCount := 5
	if err := InitTargetConfig(map[string]int{"ts": targetCount}, "app"); err != nil {
		t.Error(err.Error())
		return
	}

	pwd, err := os.Getwd()
	if err != nil {
		t.Error(err.Error())
		return
	}

	filename := filepath.Join(pwd, "handler_test.json")
	testsMaps, err := readJSONFile(filename, "TestHandler2")
	if err != nil {
		t.Error(err.Error())
		return
	}

	for index, tt := range testsMaps {
		pp.Println(tt)

		node, err := MapToNode(tt)
		if err != nil {
			t.Error(err.Error())
			return
		}

		conf, err := NodeToStruct[InjectionConf](node)
		if err != nil {
			t.Error(err.Error())
			return
		}

		displayConfig, err := conf.GetDisplayConfig()
		if err != nil {
			t.Error(err.Error())
			return
		}

		pp.Println(displayConfig)

		name, err := conf.Create(context.Background(), index%targetCount+1, map[string]string{}, map[string]string{
			"benchmark":    "clickhouse",
			"pre_duration": "1",
			"task_id":      "1",
			"trace_id":     "2",
			"group_id":     "3",
		})
		if err != nil {
			t.Error(err.Error())
			return
		}

		pp.Println(name)

		childNode := node.Children[strconv.Itoa(node.Value)]
		childNode.Children[strconv.Itoa(len(childNode.Children))] = &Node{
			Value: index%targetCount + 1,
		}
		pp.Println(NodeToMap(node, true))

		newConf, err := NodeToStruct[InjectionConf](node)
		if err != nil {
			t.Error(err.Error())
			return
		}

		groudtruth, err := newConf.GetGroundtruth()
		if err != nil {
			t.Error(err.Error())
			return
		}

		pp.Println(groudtruth)
	}
}

func readJSONFile(filename, key string) ([]map[string]any, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var dataMap map[string]any
	if err := json.Unmarshal(data, &dataMap); err != nil {
		return nil, err
	}

	if value, ok := dataMap[key]; ok {
		if items, ok := value.([]any); ok {
			var result []map[string]any
			for _, item := range items {
				if m, ok := item.(map[string]any); ok {
					result = append(result, m)
				}
			}

			return result, nil
		}
	}

	return nil, fmt.Errorf("failed to read the value of key %s", key)
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
		subMap, ok := value.(map[int]any)
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
