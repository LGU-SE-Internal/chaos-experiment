package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/k0kubun/pp"
)

// 测试获取配置
func TestHandler(t *testing.T) {
	if err := InitTargetConfig(map[string]int{"ts": 6}, "app"); err != nil {
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
	targetCount := 6
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

	for _, tt := range testsMaps {
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

		name, err := conf.Create(context.Background(), "ts0", map[string]string{}, map[string]string{
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

func TestValidate(t *testing.T) {
	targetCount := 6
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
	testsMaps, err := readJSONFile(filename, "TestValidate")
	if err != nil {
		t.Error(err.Error())
		return
	}

	for _, tt := range testsMaps {
		node, err := MapToNode(tt)
		if err != nil {
			t.Error(err.Error())
			return
		}

		result, err := Validate[InjectionConf](node, "ts")
		if err != nil {
			t.Error(err.Error())
			return
		}

		pp.Println(result)
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
