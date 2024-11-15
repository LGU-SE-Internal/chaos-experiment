package handler

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
)

// 动作空间的定义
type ActionSpace struct {
	FieldName  string
	Min        int
	Max        int
	IsOptional bool
}

// 从结构体生成动作空间
func GenerateActionSpace(v interface{}) ([]ActionSpace, error) {
	typ := reflect.TypeOf(v)
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct, got %s", typ.Kind())
	}

	var actionSpace []ActionSpace
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// 获取 range tag
		rangeTag := field.Tag.Get("range")
		if rangeTag == "" {
			continue
		}
		// 获取 optional tag
		optionalTag := field.Tag.Get("optional")
		isOptional := optionalTag == "true"
		// 解析范围
		ranges := strings.Split(rangeTag, "-")
		if len(ranges) != 2 {
			return nil, fmt.Errorf("invalid range format in field %s", field.Name)
		}

		min, err := strconv.Atoi(ranges[0])
		if err != nil {
			return nil, fmt.Errorf("invalid minimum range value in field %s", field.Name)
		}
		max, err := strconv.Atoi(ranges[1])
		if err != nil {
			return nil, fmt.Errorf("invalid maximum range value in field %s", field.Name)
		}

		actionSpace = append(actionSpace, ActionSpace{
			FieldName:  field.Name,
			Min:        min,
			Max:        max,
			IsOptional: isOptional,
		})
	}

	return actionSpace, nil
}

// 验证动作是否合法
func ValidateAction(action map[string]int, actionSpace []ActionSpace) error {
	for _, space := range actionSpace {
		value, ok := action[space.FieldName]
		// 如果 action 是可选的，且未指定，则跳过检查
		if !ok {
			if space.IsOptional {
				continue
			}
			return fmt.Errorf("missing action for field %s", space.FieldName)
		}
		if value < space.Min || value > space.Max {
			return fmt.Errorf("action %s is out of range (%d-%d): %d", space.FieldName, space.Min, space.Max, value)
		}
	}
	return nil
}

// 随机生成合法动作
func generateRandomAction(actionSpace []ActionSpace) map[string]int {
	action := make(map[string]int)
	for _, space := range actionSpace {
		action[space.FieldName] = rand.Intn(space.Max-space.Min+1) + space.Min
	}
	return action
}

func ActionToStruct(tp ChaosType, action map[string]int) (interface{}, error) {
	// Retrieve the zero value of the struct corresponding to the ChaosType
	specZeroValue, ok := SpecMap[tp]
	if !ok {
		return nil, fmt.Errorf("unknown ChaosType: %v", tp)
	}

	// Get the type of the struct
	specType := reflect.TypeOf(specZeroValue)

	// Create a new instance of the struct
	specValue := reflect.New(specType).Elem()

	// Iterate over the action map and set the struct fields
	for fieldName, fieldValue := range action {
		// Get the field by name
		field := specValue.FieldByName(fieldName)
		if !field.IsValid() {
			return nil, fmt.Errorf("unknown field: %v", fieldName)
		}
		if !field.CanSet() {
			return nil, fmt.Errorf("cannot set field: %v", fieldName)
		}

		// Set the field value based on its kind
		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			field.SetInt(int64(fieldValue))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			field.SetUint(uint64(fieldValue))
		default:
			return nil, fmt.Errorf("unsupported field type: %v", field.Kind())
		}
	}

	// Return the populated struct
	return specValue.Interface(), nil
}
