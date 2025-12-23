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
	node, err := StructToNode[InjectionConf](context.Background(), SystemTrainTicket)
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

	ctx := context.Background()

	for _, tt := range testsMaps {
		pp.Println(tt)

		node, err := MapToNode(tt)
		if err != nil {
			t.Error(err.Error())
			return
		}

		conf, err := NodeToStruct[InjectionConf](ctx, node)
		if err != nil {
			t.Error(err.Error())
			return
		}

		displayConfig, err := conf.GetDisplayConfig(ctx)
		if err != nil {
			t.Error(err.Error())
			return
		}

		pp.Println(displayConfig)

		// name, err := BatchCreate(context.Background(), []InjectionConf{*conf}, "ts0", map[string]string{}, map[string]string{
		// 	"benchmark":    "clickhouse",
		// 	"pre_duration": "1",
		// 	"task_id":      "1",
		// 	"trace_id":     "2",
		// 	"group_id":     "3",
		// })
		// if err != nil {
		// 	t.Error(err.Error())
		// 	return
		// }

		// pp.Println(name)

		newConf, err := NodeToStruct[InjectionConf](ctx, node)
		if err != nil {
			t.Error(err.Error())
			return
		}

		groudtruth, err := newConf.GetGroundtruth(ctx)
		if err != nil {
			t.Error(err.Error())
			return
		}

		pp.Println(groudtruth)
	}
}

func TestValidate(t *testing.T) {
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

		result, err := Validate[InjectionConf](context.Background(), node, SystemTrainTicket)
		if err != nil {
			t.Error(err.Error())
			return
		}

		pp.Println(result)
	}
}

// Test HTTPRequestReplaceMethod display config
func TestHTTPRequestReplaceMethodDisplayConfig(t *testing.T) {
	// Test data representing HTTPRequestReplaceMethod with specific values
	testData := map[string]any{
		"name":  "InjectionConf",
		"range": []any{0, 30},
		"children": map[string]any{
			"12": map[string]any{
				"name":  "12",
				"range": []any{0, 3},
				"children": map[string]any{
					"0": map[string]any{
						"name":        "0",
						"range":       []any{1, 60},
						"children":    nil,
						"description": "Time Unit Minute",
						"value":       4,
					},
					"1": map[string]any{
						"name":        "1",
						"range":       []any{0, 0},
						"children":    nil,
						"description": "{ts: 0}",
						"value":       0,
					},
					"2": map[string]any{
						"name":        "2",
						"range":       []any{0, 67},
						"children":    nil,
						"description": "Flattened HTTP Endpoint Index",
						"value":       34,
					},
					"3": map[string]any{
						"name":        "3",
						"range":       []any{0, 6},
						"children":    nil,
						"description": "HTTP Method to replace with",
						"value":       1,
					},
				},
				"description": "",
				"value":       0,
			},
		},
		"description": "",
		"value":       12,
	}

	// Convert to node structure
	node, err := MapToNode(testData)
	if err != nil {
		t.Errorf("MapToNode failed: %v", err)
		return
	}

	// Convert to struct
	conf, err := NodeToStruct[InjectionConf](context.Background(), node)
	if err != nil {
		t.Errorf("NodeToStruct failed: %v", err)
		return
	}

	// Get display config
	displayConfig, err := conf.GetDisplayConfig(context.Background())
	if err != nil {
		t.Errorf("GetDisplayConfig failed: %v", err)
		return
	}

	t.Logf("Display Config: %+v", displayConfig)

	// Verify that replace_method shows a string method name, not an index
	if replaceMethod, ok := displayConfig["replace_method"]; ok {
		if methodName, ok := replaceMethod.(string); ok {
			// Verify it's a valid HTTP method name
			validMethods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH"}
			isValid := false
			for _, method := range validMethods {
				if methodName == method {
					isValid = true
					break
				}
			}
			if !isValid {
				t.Errorf("replace_method should be a valid HTTP method name, got: %s", methodName)
			} else {
				t.Logf("✓ replace_method correctly shows HTTP method name: %s", methodName)
			}
		} else {
			t.Errorf("replace_method should be a string, got: %T", replaceMethod)
		}
	} else {
		t.Errorf("replace_method field not found in display config")
	}

	// Verify that the method is different from the original endpoint method
	if injectionPoint, ok := displayConfig["injection_point"]; ok {
		if injectionMap, ok := injectionPoint.(map[string]any); ok {
			if originalMethod, ok := injectionMap["method"]; ok {
				if replaceMethod, ok := displayConfig["replace_method"]; ok {
					if originalMethod == replaceMethod {
						t.Errorf("replace_method (%v) should be different from original method (%v)", replaceMethod, originalMethod)
					} else {
						t.Logf("✓ replace_method (%v) is different from original method (%v)", replaceMethod, originalMethod)
					}
				}
			}
		}
	}

	pp.Println("=== Test HTTPRequestReplaceMethod Display Config ===")
	pp.Println(displayConfig)
}

func TestGetSystmeResources(t *testing.T) {
	resourceMap, err := GetSystemResourceMap(context.Background())
	if err != nil {
		t.Errorf("GetAllResources failed: %v", err)
		return
	}

	pp.Println(resourceMap)
}

func TestGetChaosTypeResourceMappings(t *testing.T) {
	resourceMap, err := GetChaosTypeResourceMappings()
	if err != nil {
		t.Errorf("GetChaosResourceMap failed: %v", err)
		return
	}

	pp.Println(resourceMap)
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
