package handler

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/CUHK-SE-Group/chaos-experiment/internal/resourcelookup"
)

/*
Struct <=> Node <=> Map

any struct can be converted to a node, and then to a map
any node can be converted to a struct, and then to a map
any map can be converted to a node, and then to a struct
*/

// TODO 校验Node
type Node struct {
	Name        string           `json:"name"`
	Range       []int            `json:"range"`
	Children    map[string]*Node `json:"children,omitempty"`
	Description string           `json:"description,omitempty"`
	Value       int              `json:"value,omitempty"`
}

func NodeToMap(n *Node, excludeUnset bool) map[string]any {
	result := make(map[string]any)
	if excludeUnset {
		if n.Name != "" {
			result["name"] = n.Name
		}

		if n.Range != nil {
			result["range"] = n.Range
		}
	} else {
		result["name"] = n.Name
		result["range"] = n.Range
	}

	result["value"] = n.Value

	if n.Description != "" {
		result["description"] = n.Description
	}

	if len(n.Children) > 0 {
		childrenMap := make(map[string]any)
		for k, v := range n.Children {
			childrenMap[k] = NodeToMap(v, excludeUnset)
		}

		result["children"] = childrenMap
	}

	return result
}

func MapToNode(m map[string]any) (*Node, error) {
	node := &Node{}

	value, valueOK := parseValueToInt(m["value"])
	if valueOK {
		node.Value = value
	}

	children, childrenOK := m["children"].(map[string]any)
	if childrenOK {
		node.Children = make(map[string]*Node)
		for key, val := range children {
			childMap, ok := val.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid child node at key [%s]", key)
			}

			childNode, err := MapToNode(childMap)
			if err != nil {
				return nil, err
			}

			node.Children[key] = childNode
		}
	}

	if !valueOK && !childrenOK {
		return nil, fmt.Errorf("a node must contain at least one key of 'value' or 'children'")
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

	node.Children = make(map[string]*Node)

	if rt.Kind() == reflect.Struct {
		for i := range rt.NumField() {
			field := rt.Field(i)
			if child, err := buildFieldNode(field); err != nil {
				return nil, err
			} else {
				node.Children[strconv.Itoa(i)] = child
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

	if rt.Name() != "InjectionConf" && rt.PkgPath() == "github.com/CUHK-SE-Group/chaos-experiment/handler" {
		if len(n.Children) != 1 {
			return nil, fmt.Errorf("injection conf must have only one chaos type")
		}
	}

	for key := range n.Children {
		intKey, err := strconv.Atoi(key)
		if err != nil {
			return nil, err
		}

		if intKey < 0 || intKey >= rt.NumField() {
			return nil, fmt.Errorf("invalid key in the children of node")
		}

		if err := processStructField(rt.Field(intKey), val.Field(intKey), n.Children[key]); err != nil {
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

	return assignBasicType(field, val, node)
}

func processNestedStruct(rt reflect.Type, val reflect.Value, node *Node) error {
	if rt.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct type, got %s", rt.Kind())
	}

	for key := range node.Children {
		intKey, err := strconv.Atoi(key)
		if err != nil {
			return err
		}

		if intKey < 0 || intKey >= rt.NumField() {
			return fmt.Errorf("invalid key in the children of node")
		}

		if err := processStructField(rt.Field(intKey), val.Field(intKey), node.Children[key]); err != nil {
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
		case "AppName", KeyApp:
			values, err := client.GetLabels(TargetNamespace, TargetLabelKey)
			if err != nil {
				return 0, 0, fmt.Errorf("failed to get labels: %w", err)
			}
			start = 0
			end = len(values) - 1
		case KeyMethod:
			// For flattened JVM methods
			methods, err := resourcelookup.GetAllJVMMethods()
			if err != nil {
				return 0, 0, fmt.Errorf("failed to get JVM methods: %w", err)
			}
			start = 0
			end = len(methods) - 1
		case KeyEndpoint:
			// For flattened HTTP endpoints
			endpoints, err := resourcelookup.GetAllHTTPEndpoints()
			if err != nil {
				return 0, 0, fmt.Errorf("failed to get HTTP endpoints: %w", err)
			}
			start = 0
			end = len(endpoints) - 1
		case KeyNetworkPair:
			// For flattened network pairs
			pairs, err := resourcelookup.GetAllNetworkPairs()
			if err != nil {
				return 0, 0, fmt.Errorf("failed to get network pairs: %w", err)
			}
			start = 0
			end = len(pairs) - 1
		case "ContainerIdx":
			// For flattened containers
			containers, err := resourcelookup.GetAllContainers()
			if err != nil {
				return 0, 0, fmt.Errorf("failed to get containers: %w", err)
			}
			start = 0
			end = len(containers) - 1
		case KeyDNSEndpoint:
			// For flattened DNS endpoints
			endpoints, err := resourcelookup.GetAllDNSEndpoints()
			if err != nil {
				return 0, 0, fmt.Errorf("failed to get DNS endpoints: %w", err)
			}
			start = 0
			end = len(endpoints) - 1
		case KeyDatabase:
			// For flattened database operations
			dbOps, err := resourcelookup.GetAllDatabaseOperations()
			if err != nil {
				return 0, 0, fmt.Errorf("failed to get database operations: %w", err)
			}
			start = 0
			end = len(dbOps) - 1
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

func parseValueToInt(value any) (int, bool) {
	valFloat, ok := value.(float64)
	if ok {
		return int(valFloat), true
	}

	val, ok := value.(int)
	if ok {
		return val, true
	}

	return 0, false
}
