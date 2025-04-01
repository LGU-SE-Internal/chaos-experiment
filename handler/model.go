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
	Name        string        `json:"name"`
	Range       []int         `json:"range"`
	Children    map[int]*Node `json:"children,omitempty"`
	Description string        `json:"description,omitempty"`
	Value       int           `json:"value,omitempty"`
}

func NodeToMap(n *Node) map[string]any {
	result := make(map[string]any)
	result["name"] = n.Name
	result["range"] = n.Range
	result["value"] = n.Value

	if n.Description != "" {
		result["description"] = n.Description
	}

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
	start, end, err := getValueRange(field)
	if err != nil {
		return nil, err
	}

	child := &Node{
		Name:        field.Name,
		Description: field.Tag.Get("description"),
		Range:       []int{start, end},
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

	val := reflect.New(rt).Elem()

	if rt.Name() == "InjectionConf" && rt.PkgPath() == "github.com/CUHK-SE-Group/chaos-experiment/handler" {
		if len(n.Children) != 1 {
			return nil, fmt.Errorf("injection conf must have only one chaos type")
		}

		var key int
		for key = range n.Children {
		}

		if key < 0 || key >= rt.NumField() {
			return nil, fmt.Errorf("invalid key in the children of injection conf")
		}

		val := reflect.New(rt).Elem()
		if err := processStructField(rt.Field(key), val.Field(key), n.Children[key]); err != nil {
			return nil, err
		}

		return val.Addr().Interface().(*T), nil
	}

	for i := range rt.NumField() {
		if err := processStructField(rt.Field(i), val.Field(i), n.Children[i]); err != nil {
			return nil, err
		}
	}

	return val.Addr().Interface().(*T), nil
}

func processStructField(field reflect.StructField, val reflect.Value, node *Node) error {
	if node == nil {
		if field.Tag.Get("optional") == "true" {
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

	return assignBasicType(field, val, node) // 传入整个node而非range
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

func typeName(rt reflect.Type, fieldName string) string {
	if fieldName != "" {
		return fieldName
	}
	return rt.Name()
}

func assignBasicType(field reflect.StructField, val reflect.Value, node *Node) error {
	start, end, err := getValueRange(field)
	if err != nil {
		return err
	}

	if node.Value < start || node.Value > end {
		return fmt.Errorf("value %d out of range [%d, %d]", node.Value, start, end)
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

func getValueRange(field reflect.StructField) (int, int, error) {
	start, end, err := parseRangeTag(field.Tag.Get("range"))
	if err != nil {
		return 0, 0, fmt.Errorf("field %s: %w", field.Name, err)
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
				return 0, 0, fmt.Errorf("failed to get labels: %w", err)
			}
			start = 0
			end = len(values) - 1
		}
	}

	return start, end, err
}

func parseRangeTag(tag string) (int, int, error) {
	if tag == "" {
		return 0, 0, fmt.Errorf("empty range tag")
	}

	// Special handling for ranges with negative numbers
	var parts []string
	if strings.HasPrefix(tag, "-") {
		// Handle case like "-600-600"
		remainingPart := tag[1:] // Remove the first "-"
		idx := strings.Index(remainingPart, "-")
		if idx == -1 {
			return 0, 0, fmt.Errorf("invalid range format: missing second bound")
		}

		firstPart := "-" + remainingPart[:idx]
		secondPart := remainingPart[idx+1:]
		parts = []string{firstPart, secondPart}
	} else {
		// Standard case like "0-100"
		parts = strings.Split(tag, "-")
	}

	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range format: expected format 'start-end'")
	}

	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start value: %v", err)
	}

	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end value: %v", err)
	}

	return start, end, nil
}
