package handler

import (
	"fmt"
	"reflect"
	"sort"
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

var (
	NodeNsPrefixMap map[*Node]string
)

func NodeToMap(n *Node, excludeUnset bool) map[string]any {
	result := make(map[string]any)
	if excludeUnset {
		if n.Name != "" {
			result["name"] = n.Name
		}

		if n.Range != nil {
			result["range"] = n.Range
		}

		if n.Value != ValueNotSet {
			result["value"] = n.Value
		}
	} else {
		result["name"] = n.Name
		result["range"] = n.Range
		result["value"] = n.Value
	}

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

func StructToNode[T any](namespacePrefix string) (*Node, error) {
	var t T
	rt := reflect.TypeOf(t)
	if rt.Kind() != reflect.Struct {
		return nil, fmt.Errorf("struct T must be a struct type")
	}

	rootNode := &Node{}
	if NodeNsPrefixMap == nil {
		NodeNsPrefixMap = make(map[*Node]string)
	}

	NodeNsPrefixMap[rootNode] = namespacePrefix
	return buildNode(rt, "", rootNode)
}

func buildNode(rt reflect.Type, fieldName string, rootNode *Node) (*Node, error) {
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
			if field.Name == KeyNamespaceTarget {
				continue
			}

			child, err := buildFieldNode(field, rootNode)
			if err != nil {
				return nil, err
			}

			node.Children[strconv.Itoa(i)] = child
		}
	}

	return node, nil
}

func buildFieldNode(field reflect.StructField, rootNode *Node) (*Node, error) {
	start, end, err := getValueRange(field, rootNode)
	if err != nil {
		return nil, err
	}

	value := ValueNotSet
	description := field.Tag.Get("description")
	if field.Name == KeyNamespace {
		namespacePrefixMap := make(map[string]int, len(NamespacePrefixs))
		for idx, ns := range NamespacePrefixs {
			namespacePrefixMap[ns] = idx
		}

		description = mapToString(namespacePrefixMap)
		value = namespacePrefixMap[NodeNsPrefixMap[rootNode]]
	}

	child := &Node{
		Name:        field.Name,
		Description: description,
		Range:       []int{start, end},
		Value:       value,
	}

	fieldType := field.Type
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	if fieldType.Kind() == reflect.Struct {
		if nested, err := buildNode(fieldType, field.Name, rootNode); err != nil {
			return nil, err
		} else {
			child.Children = nested.Children
		}
	}

	return child, nil
}

func mapToString(m map[string]int) string {
	pairs := make([]string, 0, len(m))
	for key, value := range m {
		pairs = append(pairs, fmt.Sprintf("%s: %d", key, value))
	}
	return "{" + strings.Join(pairs, ", ") + "}"
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

		if NodeNsPrefixMap == nil {
			NodeNsPrefixMap = make(map[*Node]string)
		}

		intKey := n.Value
		if intKey < 0 || intKey >= rt.NumField() {
			return nil, fmt.Errorf("invalid key in the children of node")
		}

		if err := processStructField(rt.Field(intKey), val.Field(intKey), n.Children[strconv.Itoa(n.Value)], n); err != nil {
			return nil, err
		}
	}

	return val.Addr().Interface().(*T), nil
}

func processStructField(field reflect.StructField, val reflect.Value, node, rootNode *Node) error {
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
		_, end, err := parseRangeTag(field.Tag.Get("range"))
		if err != nil {
			return fmt.Errorf("field %s: %v", field.Name, err)
		}

		return processNestedStruct(fieldType, val, node, rootNode, end)
	}

	return assignBasicType(field, val, node, rootNode)
}

func processNestedStruct(rt reflect.Type, val reflect.Value, node, rootNode *Node, maxNum int) error {
	if rt.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct type, got %s", rt.Kind())
	}

	intKeys := make([]int, 0, len(node.Children))
	for key := range node.Children {
		intKey, err := strconv.Atoi(key)
		if err != nil {
			return err
		}

		intKeys = append(intKeys, intKey)
	}

	sort.Ints(intKeys)
	for intKey := range intKeys {
		// 对超出range的部份忽略
		if maxNum < intKey && intKey < rt.NumField() {
			continue
		}

		if intKey < 0 || intKey >= rt.NumField() {
			return fmt.Errorf("invalid key in the children of node")
		}

		if err := processStructField(rt.Field(intKey), val.Field(intKey), node.Children[strconv.Itoa(intKey)], rootNode); err != nil {
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

func assignBasicType(field reflect.StructField, val reflect.Value, node, rootNode *Node) error {
	start, end, err := getValueRange(field, rootNode)
	if err != nil {
		return err
	}

	if node.Value < start || node.Value > end {
		return fmt.Errorf("value %d out of range [%d, %d]", node.Value, start, end)
	}

	if field.Name == KeyNamespace {
		NodeNsPrefixMap[rootNode] = NamespacePrefixs[node.Value]
	}

	return setValue(val, node.Value)
}

func getValueRange(field reflect.StructField, rootNode *Node) (int, int, error) {
	start, end, err := parseRangeTag(field.Tag.Get("range"))
	if err != nil {
		return 0, 0, fmt.Errorf("field %s: %w", field.Name, err)
	}

	dyn := field.Tag.Get("dynamic")
	if dyn == "true" {
		switch field.Name {
		case KeyNamespace:
			start = 0
			end = len(NamespacePrefixs) - 1
		case KeyNamespaceTarget:
			prefix, ok := NodeNsPrefixMap[rootNode]
			if !ok {
				return 0, 0, fmt.Errorf("failed to get namespace prefix in %s", KeyNamespaceTarget)
			}

			targetCount, ok := NamespaceTargetMap[prefix]
			if !ok {
				return 0, 0, fmt.Errorf("failed to get namespace targe count")
			}

			start = DefaultStartIndex
			end = targetCount
		case KeyApp:
			prefix, ok := NodeNsPrefixMap[rootNode]
			if !ok {
				return 0, 0, fmt.Errorf("failed to get namespace prefix in %s", KeyApp)
			}

			namespace := fmt.Sprintf("%s%d", prefix, DefaultStartIndex)
			values, err := client.GetLabels(namespace, TargetLabelKey)
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
		case KeyContainer:
			// For flattened containers
			prefix, ok := NodeNsPrefixMap[rootNode]
			if !ok {
				return 0, 0, fmt.Errorf("failed to get namespace prefix in %s", KeyContainer)
			}

			namespace := fmt.Sprintf("%s%d", prefix, DefaultStartIndex)
			containers, err := resourcelookup.GetAllContainers(namespace)
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

func setValue(val reflect.Value, value int) error {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val.OverflowInt(int64(value)) {
			return fmt.Errorf("value %d overflow for type %s", value, val.Type())
		}
		val.SetInt(int64(value))

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if value < 0 {
			return fmt.Errorf("negative value %d for unsigned type %s", value, val.Type())
		}
		if val.OverflowUint(uint64(value)) {
			return fmt.Errorf("value %d overflow for type %s", value, val.Type())
		}
		val.SetUint(uint64(value))

	default:
		return fmt.Errorf("unsupported type %s", val.Kind())
	}

	return nil
}
