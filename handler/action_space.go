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

func ActionToStruct(action map[string]int, target interface{}) error {
	// 检查 target 是否为指针类型的结构体
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to a struct, currently: %v, %v", val.Kind(), val.Elem().Kind())
	}

	// 获取结构体类型和值
	val = val.Elem()
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// 检查字段是否可设置
		if !fieldValue.CanSet() {
			continue
		}

		// 检查是否有对应的值
		if actionValue, ok := action[field.Name]; ok {
			// 确保字段是 int 类型
			if fieldValue.Kind() == reflect.Int {
				fieldValue.SetInt(int64(actionValue))
			} else {
				return fmt.Errorf("field %s is not of type int", field.Name)
			}
		}
	}

	return nil
}
