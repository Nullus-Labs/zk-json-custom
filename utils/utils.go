package utils

import (
	"fmt"
	"math"
	"sample/circuit"

	"reflect"
	"strings"

	"github.com/consensys/gnark/frontend"
)

type Profile struct {
	Profiles []map[string]interface{}
}

func EmptyInteger(maxDigit ...int) circuit.Integer {
	if len(maxDigit) > 0 {
		return circuit.Integer{
			X:        0,
			MaxDigit: maxDigit[0]}
	} else {
		return circuit.Integer{
			X:        0,
			MaxDigit: 0}
	}
}

func MakeInteger(x int64, maxDigit int) circuit.Integer {
	if x >= int64(math.Pow10(maxDigit)) {
		panic("Integer is larger than maxDigit")
	}
	return circuit.Integer{
		X:        frontend.Variable(x),
		MaxDigit: maxDigit}
}

func EmptyString(maxLen int) circuit.String {
	ret := make(circuit.String, maxLen+1)
	ret[0] = 0
	for i := 1; i < maxLen+1; i++ {
		ret[i] = circuit.DUMMY
	}
	return ret
}

func MakeString(input string, maxLen int) circuit.String {
	fmt.Println(input, maxLen)
	ascii := circuit.StringToAscii(input)
	x := make(circuit.String, maxLen+1)
	x[0] = len(ascii)
	for i := 1; i < len(ascii)+1; i++ {
		x[i] = ascii[i-1]
	}
	for i := len(ascii) + 1; i < maxLen+1; i++ {
		x[i] = circuit.DUMMY
	}
	return x
}

func convertToInt(someInterface interface{}) int {
	switch value := someInterface.(type) {
	case int:
		return value
	case float64:
		// Convert float64 to int, possibly rounding down.
		return int(value)
	default:
		return 0
	}
}

func safeConvert(sliceInterface interface{}) ([]map[string]interface{}, error) {
	slice, ok := sliceInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("input is not a []interface{}")
	}

	var result []map[string]interface{}
	for _, item := range slice {
		if mapItem, ok := item.(map[string]interface{}); ok {
			result = append(result, mapItem)
		} else {
			return nil, fmt.Errorf("element is not a map[string]interface{}")
		}
	}
	return result, nil
}

func getMaxLen(name string, profile []map[string]interface{}) int {
	length := 0
	for _, p := range profile {
		if p["key"] == name {
			length = convertToInt(p["maximumLength"])
		}
		if p["valueType"] == "Array" {
			value, err := safeConvert(p["elementType"])
			if err != nil {
				panic(err)
			}
			l := getMaxLen(name, value)
			if l > length {
				length = l
			}
		} else if p["valueType"] == "Dictionary" {
			value, err := safeConvert(p["elementValueType"])
			if err != nil {
				panic(err)
			}
			l := getMaxLen(name, value)
			if l > length {
				length = l
			}
			//return getMaxLen(name, value)
			//return getMaxLen(name, p["elementValueType"].([]map[string]interface{}))
		}
	}
	return length

}

func findFieldPath(structType reflect.Type, fieldName string, path string) (string, bool) {
	// Ensure we are dealing with a struct or a pointer to a struct
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		return "", false
	}

	// Iterate over each field in the struct
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		currentPath := field.Name

		if strings.ToLower(currentPath) == strings.ToLower(fieldName) { // Check if the current field is the target field
			if path == "" {
				return currentPath, true
			} else {
				return path + "\\" + currentPath, true
			}
		}

		// If the field is a struct or a slice of structs, recurse into it
		if field.Type.Kind() == reflect.Struct || (field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct) {
			nextPath := path
			if nextPath == "" {
				nextPath = currentPath
			} else {
				nextPath += "\\" + currentPath
			}

			if field.Type.Kind() == reflect.Slice {
				field.Type = field.Type.Elem() // Handle slice of structs by getting the element type
			}

			if foundPath, found := findFieldPath(field.Type, fieldName, nextPath); found {
				return foundPath, true
			}
		}
	}

	return "", false
}

func makeEmptyStruct(s interface{}, rawProfile []map[string]interface{}) {
	sVal := reflect.ValueOf(s)
	if sVal.Kind() == reflect.Ptr {
		sVal = sVal.Elem()
	}
	profileType := reflect.TypeOf(circuit.Profile{})

	// Iterate over each field in the struct.
	for i := 0; i < sVal.NumField(); i++ {
		field := sVal.Field(i)

		// Ensure the field is settable.
		if field.CanSet() {
			switch field.Type() {
			case reflect.TypeOf(circuit.Integer{}):
				name, _ := findFieldPath(profileType, sVal.Type().Field(i).Name, "")
				maxLen := getMaxLen(name, rawProfile)
				field.Set(reflect.ValueOf(EmptyInteger(maxLen)))
			case reflect.TypeOf(circuit.String{}):
				name, _ := findFieldPath(profileType, sVal.Type().Field(i).Name, "")
				maxLen := getMaxLen(name, rawProfile)
				field.Set(reflect.ValueOf(EmptyString(maxLen)))
			default:
				if field.Kind() == reflect.Struct {
					// Recursively initialize nested structs.
					makeEmptyStruct(field.Addr().Interface(), rawProfile)
				}
			}
		}
	}
}

//func makeEmptyStructOld(s interface{}, rawProfile []map[string]interface{}) {
//	sVal := reflect.ValueOf(s)
//
//	if circuit.CheckInterface(sVal.Interface()) {
//		sVal.Set(sVal.Interface().Init())
//	}
//
//	if sVal.Kind() == reflect.Ptr || sVal.Kind() == reflect.Interface {
//		sVal = sVal.Elem()
//	}
//	profileType := reflect.TypeOf(circuit.Profile{})
//	for i := 0; i < sVal.NumField(); i++ {
//		sField := sVal.Field(i)
//		switch sField.Kind() {
//		case reflect.Int64:
//			if sField.Type() == reflect.TypeOf(circuit.Integer{}) {
//				name, _ := findFieldPath(profileType, sVal.Type().Field(i).Name, "")
//				maxLen := getMaxLen(name, rawProfile)
//				sField.Set(reflect.ValueOf(EmptyInteger(maxLen)))
//			}
//		case reflect.String:
//			if sField.Type() == reflect.TypeOf(circuit.String{}) {
//				name, _ := findFieldPath(profileType, sVal.Type().Field(i).Name, "")
//				maxLen := getMaxLen(name, rawProfile)
//				sField.Set(reflect.ValueOf(EmptyStringNormal(maxLen)))
//			}
//		case reflect.Slice:
//			// Initialize and fill slices
//			name, _ := findFieldPath(profileType, sVal.Type().Field(i).Name, "")
//			maxLen := getMaxLen(name, rawProfile)
//			elementType := sField.Type().Elem()
//			newSlice := reflect.MakeSlice(sField.Type(), maxLen, maxLen)
//			for j := 0; j < sField.Len(); j++ {
//				newElem := reflect.New(elementType)
//				makeEmptyStruct(newElem.Interface(), rawProfile)
//				//transferValues(s1Field.Index(j).Addr().Interface(), newElem.Interface(), rawProfile)
//				newSlice.Index(j).Set(newElem.Elem())
//			}
//			sField.Set(newSlice)
//
//		case reflect.Struct:
//			// Recursive handling for nested structs
//			if sField.CanAddr() && sField.CanSet() {
//				makeEmptyStruct(sField.Addr().Interface(), rawProfile)
//			}
//		default:
//			panic("Unsupported type")
//		}
//
//	}
//}

func transferValues(s1 interface{}, s2 interface{}, rawProfile []map[string]interface{}) {

	s1Val := reflect.ValueOf(s1)
	if s1Val.Kind() == reflect.Ptr || s1Val.Kind() == reflect.Interface {
		s1Val = s1Val.Elem()
	}
	s2Val := reflect.ValueOf(s2)
	if s2Val.Kind() == reflect.Ptr || s2Val.Kind() == reflect.Interface {
		s2Val = s2Val.Elem()
	}

	profileType := reflect.TypeOf(circuit.Profile{})

	for i := 0; i < s1Val.NumField(); i++ {
		s1Field := s1Val.Field(i)
		s2Field := s2Val.Field(i)

		// Type switch to handle various conversions
		switch s1Field.Kind() {
		case reflect.Int64:
			if s2Field.Type() == reflect.TypeOf(circuit.Integer{}) {
				name, _ := findFieldPath(profileType, s1Val.Type().Field(i).Name, "")
				maxLen := getMaxLen(name, rawProfile)
				//maxLen := getMaxLen(findFieldPath( , s1Val.Type().Field(i).Name), profile)
				s2Field.Set(reflect.ValueOf(MakeInteger(s1Field.Int(), maxLen)))
			}
		case reflect.String:
			if s2Field.Type() == reflect.TypeOf(circuit.String{}) {
				name, _ := findFieldPath(profileType, s1Val.Type().Field(i).Name, "")
				maxLen := getMaxLen(name, rawProfile)
				s2Field.Set(reflect.ValueOf(MakeString(s1Field.String(), maxLen)))
			}
		case reflect.Slice:
			// Initialize and fill slices
			name, _ := findFieldPath(profileType, s1Val.Type().Field(i).Name, "")
			maxLen := getMaxLen(name, rawProfile)
			elementType := s2Field.Type().Elem()
			newSlice := reflect.MakeSlice(s2Field.Type(), maxLen, maxLen)
			for j := 0; j < s1Field.Len(); j++ {
				newElem := reflect.New(elementType)
				transferValues(s1Field.Index(j).Addr().Interface(), newElem.Interface(), rawProfile)
				newSlice.Index(j).Set(newElem.Elem())
			}
			if maxLen > s1Field.Len() {
				for j := s1Field.Len(); j < maxLen; j++ {
					newElem := reflect.New(elementType)
					makeEmptyStruct(newElem.Interface(), rawProfile)
					newSlice.Index(j).Set(newElem.Elem())
				}
			}
			s2Field.Set(newSlice)

		case reflect.Struct:
			// Recursive handling for nested structs
			if s1Field.CanAddr() && s2Field.CanSet() {
				transferValues(s1Field.Addr().Interface(), s2Field.Addr().Interface(), rawProfile)
			} else {
				transferValues(s1Field.Interface(), s2Field.Addr().Interface(), rawProfile)
			}
		}
	}
}

func extractFieldFromRawData(rawData []map[string]interface{}, value string, res map[string]interface{}) {
	for _, data := range rawData {
		if data["valueType"] == "Array" {
			temp, _ := safeConvert(data["elementType"])
			extractFieldFromRawData(temp, value, res)
		} else if data["valueType"] == "Dictionary" {
			temp, _ := safeConvert(data["elementValueType"])
			extractFieldFromRawData(temp, value, res)
		} else {
			if data[value] != nil {
				res[data["key"].(string)] = data[value]
			}
		}
	}
}

func valueToFrontendVariable(value interface{}, key string, rawData Profile) interface{} {
	switch value.(type) {
	case int64:
		return MakeInteger(value.(int64), getMaxLen(key, rawData.Profiles))
	case string:
		return MakeString(value.(string), getMaxLen(key, rawData.Profiles))
	case []int64:
		ret := make([]circuit.Integer, len(value.([]int64)))
		for i, v := range value.([]int64) {
			ret[i] = MakeInteger(v, getMaxLen(key, rawData.Profiles))
		}
		return ret
	case []string:
		ret := make([]circuit.String, len(value.([]string)))
		for i, v := range value.([]string) {
			ret[i] = MakeString(v, getMaxLen(key, rawData.Profiles))
		}
		return ret
	case []interface{}:
		switch value.([]interface{})[0].(type) {
		case int64:
			ret := make([]frontend.Variable, len(value.([]interface{})))
			for i, v := range value.([]interface{}) {
				ret[i] = frontend.Variable(v.(int))
			}
			return ret
		case string:
			ret := make([]circuit.String, len(value.([]interface{})))
			for i, v := range value.([]interface{}) {
				ret[i] = MakeString(v.(string), getMaxLen(key, rawData.Profiles))
			}
			return ret
		case float64:
			ret := make([]frontend.Variable, len(value.([]interface{})))
			for i, v := range value.([]interface{}) {
				ret[i] = frontend.Variable(convertToInt(v))
			}
			return ret
		default:
			panic("Unsupported type")
		}
	default:
		panic("Unsupported type")
	}
}

func makeProfile(profileJson circuit.ProfileJSON, rawData Profile) circuit.Profile {
	profile := &circuit.Profile{}
	transferValues(profileJson, profile, rawData.Profiles)
	return *profile
}

func makeLimit(rawData Profile) circuit.Limit {
	limit := circuit.Limit{}
	valLimit := reflect.ValueOf(&limit).Elem()
	flatData := make(map[string]interface{})
	extractFieldFromRawData(rawData.Profiles, "limit", flatData)
	for key, value := range flatData {
		key = ReplaceBackslashes(key)
		valLimit.FieldByName(key).Set(reflect.ValueOf(valueToFrontendVariable(value, key, rawData)))
	}
	return limit
}

func ReplaceBackslashes(input string) string {
	return strings.Replace(input, "\\", "___", -1) // -1 for replacing all occurrences
}
