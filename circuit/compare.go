package circuit

import (
	"fmt"
	"github.com/consensys/gnark/frontend"
	"reflect"
)

func checkAppendOnly(api frontend.API, rawOldContent interface{}, rawNewContent interface{}) frontend.Variable {
	//Type convert
	oldContent, err := toSliceInterface(rawOldContent)
	if err != nil {
		panic(fmt.Sprintf("oldContent should be a slice, but got %v", reflect.TypeOf(rawOldContent)))

	}

	newContent, err := toSliceInterface(rawNewContent)
	if err != nil {
		panic(fmt.Sprintf("newContent should be a slice, but got %v", reflect.TypeOf(rawNewContent)))

	}

	//Check length
	if len(oldContent) != len(newContent) {
		panic("oldContent and newContent should have the same length")
	}

	//check append only
	result := frontend.Variable(1)
	foundEmpty := frontend.Variable(0)
	for i := 0; i < len(oldContent); i++ {
		oldEmpty := structIsEmpty(api, oldContent[i])

		foundEmpty = api.Select(oldEmpty, frontend.Variable(1), foundEmpty)

		isMatch := isEqualStruct(api, oldContent[i], newContent[i])
		result = api.Select(foundEmpty, result, api.And(isMatch, result))

	}
	return result
}

func toSliceInterface(slice interface{}) ([]interface{}, error) {
	// Check if the input is actually a slice
	val := reflect.ValueOf(slice)
	if val.Kind() != reflect.Slice {
		return nil, fmt.Errorf("provided value is not a slice")
	}

	// Create a new slice of interface{} with the same length as the input slice
	result := make([]interface{}, val.Len())

	// Copy elements from the input slice to the new slice of interface{}
	for i := 0; i < val.Len(); i++ {
		result[i] = val.Index(i).Interface()
	}

	return result, nil
}

func structIsEmpty(api frontend.API, a interface{}) frontend.Variable {
	return deepIsEmpty(api, reflect.ValueOf(a))
}

func deepIsEmpty(api frontend.API, a reflect.Value) frontend.Variable {
	if CheckInterface(a.Interface()) {
		return a.Interface().(IsEmptyInterface).IsEmpty(api)
	} else {
		counter := frontend.Variable(0)
		for i := 0; i < a.NumField(); i++ {
			counter = api.Add(counter, deepIsEmpty(api, a.Field(i)))
		}
		return isEqual(api, counter, frontend.Variable(a.NumField()))
	}

}

func CheckInterface(a interface{}) bool {
	if _, ok := a.(IsEmptyInterface); ok {
		return true
	}
	return false
}

func isEqualStruct(api frontend.API, a, b interface{}) frontend.Variable {
	return deepEqualReflect(api, reflect.ValueOf(a), reflect.ValueOf(b))
}

func deepEqualReflect(api frontend.API, a, b reflect.Value) frontend.Variable {
	//print name of field
	res := isEqualInterface(api, a.Interface(), b.Interface())
	if res != 2 {
		return res
	}
	counter := frontend.Variable(0)
	for i := 0; i < a.NumField(); i++ {
		innerRes := deepEqualReflect(api, a.Field(i), b.Field(i))
		counter = api.Add(counter, innerRes)
	}
	res = api.Sub(counter, frontend.Variable(a.NumField()))
	return api.IsZero(res)
}

func isEqualInterface(api frontend.API, a interface{}, b interface{}) frontend.Variable {
	if x, ok := a.(Integer); ok {
		if y, ok2 := b.(Integer); ok2 {
			return isEqualInteger(api, x, y)
		}
	} else if x, ok := a.(String); ok {
		if y, ok2 := b.(String); ok2 {
			return isEqualString(api, x, y)
		}
	} else if x, ok := a.(Array); ok {
		if y, ok2 := b.(Array); ok2 {
			return isEqualArray(api, x, y)
		}
	} else if x, ok := a.(Dict); ok {
		if y, ok2 := b.(Dict); ok2 {
			return isEqualDict(api, x, y)
		}
	} else {
		return frontend.Variable(2)
	}
	return frontend.Variable(0)
}

func isEqualDict(api frontend.API, a Dict, b Dict) frontend.Variable {
	if len(a.keys) != len(b.keys) || len(a.values) != len(b.values) {
		return 0
	}
	judge := frontend.Variable(0)
	for i := 0; i < len(a.keys); i++ {
		judge = api.Add(judge, isEqualString(api, a.keys[i], b.keys[i]))
		judge = api.Add(judge, isEqualInterface(api, a.values[i], b.values[i]))
	}
	return isEqual(api, judge, len(a.keys)*2)
}

func isEqualArray(api frontend.API, a Array, b Array) frontend.Variable {
	if len(a) != len(b) {
		return 0
	}
	judge := frontend.Variable(0)
	for i := 0; i < len(a); i++ {
		judge = api.Add(judge, isEqualInterface(api, a[i], b[i]))
	}
	return isEqual(api, judge, len(a))
}

func isEqualString(api frontend.API, x String, y String) frontend.Variable {
	if len(x) != len(y) {
		return 0
	}
	judge := frontend.Variable(0)
	for i := 0; i < len(x); i++ {
		judge = api.Add(judge, isEqual(api, x[i], y[i]))
	}
	return isEqual(api, judge, len(x))
}

func isEqualInteger(api frontend.API, a Integer, b Integer) frontend.Variable {
	return isEqual(api, a.X, b.X)
}

func checkWithinRange(api frontend.API, lower frontend.Variable, upper frontend.Variable, value Integer) frontend.Variable {
	return api.Or(api.And(isLessOrEqual(api, value.X, upper), isLessOrEqual(api, lower, value.X)), value.IsEmpty(api))
}

// todo: deal with dummy, in reality, n may be variable-length
func checkOneOfSet(api frontend.API, n int, set []String, value String) frontend.Variable {
	judge := frontend.Variable(0)
	for i := 0; i < n; i++ {
		judge = api.Add(judge, isEqualInterface(api, set[i], value))
	}
	return judge
}

func checkTimeInRange(api frontend.API, timeRange frontend.Variable, initTime frontend.Variable, targetTime frontend.Variable) frontend.Variable {
	//target time is within the time range of init time and target time is smaller than init time
	return api.And(isLess(api, initTime, targetTime), isLess(api, api.Add(initTime, timeRange), targetTime))
}

func checkFormat(api frontend.API, n int, format []frontend.Variable, value String) frontend.Variable {
	// Predefine: 1: Capital Letter, 2: Small Letter, 3: Number, 4: Special Character
	// n is the length of the format
	judge := frontend.Variable(1)
	//Skip first position
	for i := 1; i < n+1; i++ {
		check1 := api.Select(isEqual(api, format[i-1], 1), api.And(isLess(api, api.Sub(value[i], 65), 26), isGreater(api, api.Sub(value[i], 65), 0)), 0)
		check2 := api.Select(isEqual(api, format[i-1], 2), api.And(isLess(api, api.Sub(value[i], 97), 26), isGreater(api, api.Sub(value[i], 97), 0)), 0)
		check3 := api.Select(isEqual(api, format[i-1], 3), api.And(isLess(api, api.Sub(value[i], 48), 10), isGreater(api, api.Sub(value[i], 48), 0)), 0)
		check4 := api.Select(isEqual(api, format[i-1], 4), api.And(isLess(api, api.Sub(value[i], 33), 15), isGreater(api, api.Sub(value[i], 33), 0)), 0)
		judge = api.And(judge, api.Or(api.Or(check1, check2), api.Or(check3, check4)))
	}
	return api.Or(judge, value.IsEmpty(api))
}
