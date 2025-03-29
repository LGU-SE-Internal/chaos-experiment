package handler

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/CUHK-SE-Group/chaos-experiment/client"
)

/*
Struct <=> Node <=> Map

any struct can be converted to a node, and then to a map
any node can be converted to a struct, and then to a map
any map can be converted to a node, and then to a struct
*/
type Node struct {
	Name     string        `json:"name"`
	Range    []int         `json:"range"`
	Children map[int]*Node `json:"children,omitempty"`
	Value    int           `json:"value,omitempty"`
}

func NodeToMap(n *Node) map[string]any {
	result := make(map[string]any)
	result["name"] = n.Name
	result["range"] = n.Range
	result["value"] = n.Value

	if len(n.Children) > 0 {
		childrenMap := make(map[int]any)
		for k, v := range n.Children {
			childrenMap[k] = NodeToMap(v)
		}
		result["children"] = childrenMap
	}

	return result
}

func MapToNode(m map[string]any) (*Node, error) {
	node := &Node{}

	if r, ok := m["range"].([]int); ok {
		node.Range = r
	} else {
		return nil, fmt.Errorf("invalid range format")
	}

	if name, ok := m["name"].(string); ok {
		node.Name = name
	} else {
		return nil, fmt.Errorf("invalid name format, expected string")
	}

	if value, ok := m["value"].(int); ok {
		node.Value = value
	}

	if children, ok := m["children"].(map[int]any); ok {
		node.Children = make(map[int]*Node)
		for key, val := range children {
			childMap, ok := val.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid child node at key %d", key)
			}
			childNode, err := MapToNode(childMap)
			if err != nil {
				return nil, err
			}
			node.Children[key] = childNode
		}
	}

	return node, nil
}

func StructToNode[T any]() (*Node, error) {
	var t T
	rt := reflect.TypeOf(t)
	if rt.Kind() != reflect.Struct {
		return nil, fmt.Errorf("struct T must be a struct type")
	}

	return buildNode(rt, "")
}

func buildNode(rt reflect.Type, fieldName string) (*Node, error) {
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	node := &Node{
		Name:  typeName(rt, fieldName),
		Range: []int{0, rt.NumField() - 1},
	}

	node.Children = make(map[int]*Node)

	if rt.Kind() == reflect.Struct {
		for i := range rt.NumField() {
			field := rt.Field(i)
			if child, err := buildFieldNode(field); err != nil {
				return nil, err
			} else {
				node.Children[i] = child
			}
		}
	}

	return node, nil
}

func buildFieldNode(field reflect.StructField) (*Node, error) {
	start, end, err := parseRangeTag(field.Tag.Get("range"))
	if err != nil {
		return nil, fmt.Errorf("field %s: %w", field.Name, err)
	}

	dyn := field.Tag.Get("dynamic")
	if dyn == "true" {
		switch field.Name {
		case "Namespace":
			start = 0
			end = 0
		case "AppName":
			values, err := client.GetLabels(TargetNamespace, TargetLabelKey)
			if err != nil {
				return nil, fmt.Errorf("failed to get labels: %w", err)
			}
			start = 0
			end = len(values) - 1
		}
	}

	child := &Node{
		Name:  field.Name,
		Range: []int{start, end},
	}

	fieldType := field.Type
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	if fieldType.Kind() == reflect.Struct {
		if nested, err := buildNode(fieldType, field.Name); err != nil {
			return nil, err
		} else {
			child.Children = nested.Children
		}
	}

	return child, nil
}

func NodeToStruct[T any](n *Node) (*T, error) {
	var t T
	rt := reflect.TypeOf(t)
	if rt.Kind() != reflect.Struct {
		return nil, fmt.Errorf("struct T must be a struct type")
	}

	if err := validateNodeRange(rt, n); err != nil {
		return nil, err
	}

	val := reflect.New(rt).Elem()
	for i := range rt.NumField() {
		if err := processStructField(rt.Field(i), val.Field(i), n.Children[i]); err != nil {
			return nil, err
		}
	}

	return val.Addr().Interface().(*T), nil
}

func validateNodeRange(rt reflect.Type, n *Node) error {
	expected := rt.NumField() - 1
	if n.Range[1] != expected {
		return fmt.Errorf("node range mismatch: expected 0-%d, got %d-%d",
			expected, n.Range[0], n.Range[1])
	}
	return nil
}

func processStructField(field reflect.StructField, val reflect.Value, node *Node) error {
	if node == nil {
		if tagValue(field, "optional") == "true" {
			return nil
		}
		return fmt.Errorf("missing required field: %s", field.Name)
	}

	fieldType := field.Type
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
		if val.IsNil() {
			val.Set(reflect.New(fieldType))
		}
		val = val.Elem()
	}

	if fieldType.Kind() == reflect.Struct {
		return processNestedStruct(fieldType, val, node)
	}

	return assignBasicType(val, node) // 传入整个node而非range
}

func typeName(rt reflect.Type, fieldName string) string {
	if fieldName != "" {
		return fieldName
	}
	return rt.Name()
}

func parseRangeTag(tag string) (int, int, error) {
	parts := strings.Split(tag, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range format")
	}

	start, _ := strconv.Atoi(parts[0])
	end, _ := strconv.Atoi(parts[1])
	return start, end, nil
}

func tagValue(field reflect.StructField, key string) string {
	return field.Tag.Get(key)
}

func processNestedStruct(rt reflect.Type, val reflect.Value, node *Node) error {
	if rt.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct type, got %s", rt.Kind())
	}

	for i := range rt.NumField() {
		if err := processStructField(rt.Field(i), val.Field(i), node.Children[i]); err != nil {
			return err
		}
	}
	return nil
}

func assignBasicType(val reflect.Value, node *Node) error {
	if node.Value < node.Range[0] || node.Value > node.Range[1] {
		return fmt.Errorf("value %d out of range [%d, %d]", node.Value, node.Range[0], node.Range[1])
	}

	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val.OverflowInt(int64(node.Value)) {
			return fmt.Errorf("value %d overflow for type %s", node.Value, val.Type())
		}
		val.SetInt(int64(node.Value))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if node.Value < 0 {
			return fmt.Errorf("negative value %d for unsigned type %s", node.Value, val.Type())
		}
		if val.OverflowUint(uint64(node.Value)) {
			return fmt.Errorf("value %d overflow for type %s", node.Value, val.Type())
		}
		val.SetUint(uint64(node.Value))
	default:
		return fmt.Errorf("unsupported type %s", val.Kind())
	}
	return nil
}
