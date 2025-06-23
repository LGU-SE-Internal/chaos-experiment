package handler

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/LGU-SE-Internal/chaos-experiment/internal/resourcelookup"
)

/*
Struct <=> Node

any struct can be converted to a node
any node can be converted to a struct

*/

// TODO 校验Node
type Node struct {
	Name        string           `json:"name"`
	Range       []int            `json:"range"`
	Children    map[string]*Node `json:"children"`
	Description string           `json:"description"`
	Value       int              `json:"value"`
}

var (
	NodeNsPrefixMap map[*Node]string
)

func Validate[T any](target *Node, namespacePrefix string) (bool, error) {
	reference, err := StructToNode[T](namespacePrefix)
	if err != nil {
		return false, fmt.Errorf("failed to convert struct to node: %v", err)
	}

	return validateNode(reference, target, true)
}

func validateNode(reference *Node, target *Node, isFirstLevel bool) (bool, error) {
	if reference == nil || target == nil {
		return false, fmt.Errorf("one of the nodes is nil")
	}

	if target.Children == nil && reference.Range != nil && len(reference.Range) == 2 {
		if target.Value == ValueNotSet {
			return false, fmt.Errorf("target value is not set but reference has range constraint %v", reference.Range)
		}

		if target.Value < reference.Range[0] || target.Value > reference.Range[1] {
			return false, fmt.Errorf("target value %d is not within reference range %v", target.Value, reference.Range)
		}
	}

	if target.Children != nil {
		if reference.Children == nil {
			return false, fmt.Errorf("target has children but reference doesn't have children")
		}

		if !isFirstLevel && len(target.Children) != len(reference.Children) {
			return false, fmt.Errorf("children count mismatch: target has %d children, reference has %d children",
				len(target.Children), len(reference.Children))
		}

		for key, targetChild := range target.Children {
			if _, exists := reference.Children[key]; !exists {
				return false, fmt.Errorf("target child key '%s' not found in reference children", key)
			}

			return validateNode(reference.Children[key], targetChild, false)
		}
	}

	return true, nil
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
		return nil, fmt.Errorf("NodeToStruct: type parameter T must be a struct type, got %s", rt.Kind())
	}

	if n == nil {
		return nil, fmt.Errorf("NodeToStruct: input node is nil for struct type %s", rt.Name())
	}

	val := reflect.New(rt).Elem()
	if rt.Name() == "InjectionConf" && rt.PkgPath() == "github.com/LGU-SE-Internal/chaos-experiment/handler" {
		if len(n.Children) != 1 {
			childCount := len(n.Children)
			childKeys := make([]string, 0, childCount)
			for k := range n.Children {
				childKeys = append(childKeys, k)
			}
			return nil, fmt.Errorf("InjectionConf must have exactly one chaos type, got %d children with keys: %v", childCount, childKeys)
		}

		if NodeNsPrefixMap == nil {
			NodeNsPrefixMap = make(map[*Node]string)
		}

		intKey := n.Value
		if intKey < 0 || intKey >= rt.NumField() {
			return nil, fmt.Errorf("invalid field index %d for struct %s (valid range: 0-%d)", intKey, rt.Name(), rt.NumField()-1)
		}

		expectedChildKey := strconv.Itoa(n.Value)
		childNode, exists := n.Children[expectedChildKey]
		if !exists {
			availableKeys := make([]string, 0, len(n.Children))
			for k := range n.Children {
				availableKeys = append(availableKeys, k)
			}
			return nil, fmt.Errorf("expected child key '%s' not found in node children, available keys: %v", expectedChildKey, availableKeys)
		}

		fieldName := rt.Field(intKey).Name
		if err := processStructField(rt.Field(intKey), val.Field(intKey), childNode, n); err != nil {
			return nil, fmt.Errorf("failed to process field '%s' (index %d) in struct %s: %w", fieldName, intKey, rt.Name(), err)
		}
	}

	return val.Addr().Interface().(*T), nil
}

func processStructField(field reflect.StructField, val reflect.Value, node, rootNode *Node) error {
	if node == nil {
		if field.Tag.Get("optional") == "true" {
			return nil
		}

		return fmt.Errorf("missing required field '%s' (type: %s)", field.Name, field.Type)
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
			return fmt.Errorf("field '%s' (type: %s) has invalid range tag: %w", field.Name, field.Type, err)
		}

		if err := processNestedStruct(fieldType, val, node, rootNode, end); err != nil {
			return fmt.Errorf("failed to process nested struct field '%s' (type: %s): %w", field.Name, fieldType.Name(), err)
		}
		return nil
	}

	if err := assignBasicType(field, val, node, rootNode); err != nil {
		return fmt.Errorf("failed to assign value to field '%s' (type: %s): %w", field.Name, field.Type, err)
	}
	return nil
}

func processNestedStruct(rt reflect.Type, val reflect.Value, node, rootNode *Node, maxNum int) error {
	if rt.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct type, got %s for type %s", rt.Kind(), rt.Name())
	}

	if node.Children == nil {
		return fmt.Errorf("node has no children for struct type %s", rt.Name())
	}

	intKeys := make([]int, 0, len(node.Children))
	invalidKeys := make([]string, 0)

	for key := range node.Children {
		intKey, err := strconv.Atoi(key)
		if err != nil {
			invalidKeys = append(invalidKeys, key)
			continue
		}
		intKeys = append(intKeys, intKey)
	}

	if len(invalidKeys) > 0 {
		return fmt.Errorf("struct %s contains non-integer child keys: %v (all keys must be numeric field indices)", rt.Name(), invalidKeys)
	}

	sort.Ints(intKeys)
	for _, intKey := range intKeys {
		// 对超出range的部份忽略
		if maxNum < intKey && intKey < rt.NumField() {
			continue
		}

		if intKey < 0 || intKey >= rt.NumField() {
			return fmt.Errorf("invalid field index %d for struct %s (valid range: 0-%d), available field count: %d",
				intKey, rt.Name(), rt.NumField()-1, rt.NumField())
		}

		field := rt.Field(intKey)
		childNode := node.Children[strconv.Itoa(intKey)]

		if err := processStructField(field, val.Field(intKey), childNode, rootNode); err != nil {
			return fmt.Errorf("failed to process field '%s' (index %d) in struct %s: %w",
				field.Name, intKey, rt.Name(), err)
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
		return fmt.Errorf("failed to get value range for field '%s': %w", field.Name, err)
	}

	if node.Value < start || node.Value > end {
		return fmt.Errorf("field '%s': value %d is out of valid range [%d, %d]",
			field.Name, node.Value, start, end)
	}

	if field.Name == KeyNamespace {
		if node.Value >= len(NamespacePrefixs) {
			return fmt.Errorf("field '%s': namespace index %d exceeds available namespaces count %d",
				field.Name, node.Value, len(NamespacePrefixs))
		}
		NodeNsPrefixMap[rootNode] = NamespacePrefixs[node.Value]
	}

	if err := setValue(val, node.Value); err != nil {
		return fmt.Errorf("field '%s': failed to set value %d: %w", field.Name, node.Value, err)
	}

	return nil
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
			start = DefaultStartIndex
			end = len(NamespacePrefixs) - 1
		case KeyApp:
			prefix, ok := NodeNsPrefixMap[rootNode]
			if !ok {
				return 0, 0, fmt.Errorf("failed to get namespace prefix in %s", KeyApp)
			}

			namespace := fmt.Sprintf("%s%d", prefix, DefaultStartIndex)
			values, err := resourcelookup.GetAllAppLabels(namespace, TargetLabelKey)
			if err != nil || len(values) == 0 {
				return 0, 0, fmt.Errorf("failed to get labels: %w", err)
			}

			start = DefaultStartIndex
			end = len(values) - 1
		case KeyMethod:
			// For flattened JVM methods
			methods, err := resourcelookup.GetAllJVMMethods()
			if err != nil {
				return 0, 0, fmt.Errorf("failed to get JVM methods: %w", err)
			}

			start = DefaultStartIndex
			end = len(methods) - 1
		case KeyEndpoint:
			// For flattened HTTP endpoints
			endpoints, err := resourcelookup.GetAllHTTPEndpoints()
			if err != nil {
				return 0, 0, fmt.Errorf("failed to get HTTP endpoints: %w", err)
			}

			start = DefaultStartIndex
			end = len(endpoints) - 1
		case KeyNetworkPair:
			// For flattened network pairs
			pairs, err := resourcelookup.GetAllNetworkPairs()
			if err != nil {
				return 0, 0, fmt.Errorf("failed to get network pairs: %w", err)
			}

			start = DefaultStartIndex
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

			start = DefaultStartIndex
			end = len(containers) - 1
		case KeyDNSEndpoint:
			// For flattened DNS endpoints
			endpoints, err := resourcelookup.GetAllDNSEndpoints()
			if err != nil {
				return 0, 0, fmt.Errorf("failed to get DNS endpoints: %w", err)
			}

			start = DefaultStartIndex
			end = len(endpoints) - 1
		case KeyDatabase:
			// For flattened database operations
			dbOps, err := resourcelookup.GetAllDatabaseOperations()
			if err != nil {
				return 0, 0, fmt.Errorf("failed to get database operations: %w", err)
			}

			start = DefaultStartIndex
			end = len(dbOps) - 1
		}
	}

	return start, end, err
}

func parseRangeTag(tag string) (int, int, error) {
	if tag == "" {
		return 0, 0, fmt.Errorf("range tag is empty (expected format: 'start-end', e.g., '0-100')")
	}

	// Special handling for ranges with negative numbers
	var parts []string
	if strings.HasPrefix(tag, "-") {
		// Handle case like "-600-600"
		remainingPart := tag[1:] // Remove the first "-"
		idx := strings.Index(remainingPart, "-")
		if idx == -1 {
			return 0, 0, fmt.Errorf("invalid range format '%s': missing second bound (expected format: 'start-end')", tag)
		}

		firstPart := "-" + remainingPart[:idx]
		secondPart := remainingPart[idx+1:]
		parts = []string{firstPart, secondPart}
	} else {
		// Standard case like "0-100"
		parts = strings.Split(tag, "-")
	}

	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range format '%s': expected format 'start-end' (e.g., '0-100' or '-50-50'), got %d parts", tag, len(parts))
	}

	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start value '%s' in range '%s': %v", parts[0], tag, err)
	}

	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end value '%s' in range '%s': %v", parts[1], tag, err)
	}

	if start > end {
		return 0, 0, fmt.Errorf("invalid range '%s': start value %d is greater than end value %d", tag, start, end)
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
			return fmt.Errorf("value %d causes overflow for %s type (max: %d)",
				value, val.Type(), getMaxValueForType(val.Type()))
		}
		val.SetInt(int64(value))

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if value < 0 {
			return fmt.Errorf("cannot assign negative value %d to unsigned type %s", value, val.Type())
		}
		if val.OverflowUint(uint64(value)) {
			return fmt.Errorf("value %d causes overflow for %s type (max: %d)",
				value, val.Type(), getMaxValueForUnsignedType(val.Type()))
		}
		val.SetUint(uint64(value))

	default:
		return fmt.Errorf("unsupported type %s for integer assignment (supported: int, uint and their variants)", val.Kind())
	}

	return nil
}

// Helper functions to get max values for better error messages
func getMaxValueForType(t reflect.Type) int64 {
	switch t.Kind() {
	case reflect.Int8:
		return 127
	case reflect.Int16:
		return 32767
	case reflect.Int32:
		return 2147483647
	case reflect.Int64, reflect.Int:
		return 9223372036854775807
	default:
		return 0
	}
}

func getMaxValueForUnsignedType(t reflect.Type) uint64 {
	switch t.Kind() {
	case reflect.Uint8:
		return 255
	case reflect.Uint16:
		return 65535
	case reflect.Uint32:
		return 4294967295
	case reflect.Uint64, reflect.Uint:
		return 18446744073709551615
	default:
		return 0
	}
}
