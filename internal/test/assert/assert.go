package assert

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func formatMsg(msgAndArgs []any) string {
	if len(msgAndArgs) == 0 {
		return ""
	}
	if len(msgAndArgs) == 1 {
		if msg, ok := msgAndArgs[0].(string); ok {
			return "\n" + msg
		}
		return fmt.Sprintf("\n%v", msgAndArgs[0])
	}
	if msg, ok := msgAndArgs[0].(string); ok {
		return "\n" + fmt.Sprintf(msg, msgAndArgs[1:]...)
	}
	return "\n" + fmt.Sprint(msgAndArgs...)
}

func Contains(t *testing.T, s, contains any, msgAndArgs ...any) bool {
	t.Helper()
	ok, found := includeElement(s, contains)
	if !ok {
		t.Errorf("Contains could not be applied to type %T%s", s, formatMsg(msgAndArgs))
		return false
	}
	if !found {
		t.Errorf("%#v does not contain %#v%s", s, contains, formatMsg(msgAndArgs))
		return false
	}
	return true
}

// includeElement is a helper for Contains
func includeElement(list any, element any) (ok, found bool) {
	listValue := reflect.ValueOf(list)
	defer func() {
		if e := recover(); e != nil {
			ok = false
			found = false
		}
	}()

	if reflect.TypeOf(list) == nil {
		return false, false
	}

	if listValue.Kind() == reflect.String {
		elementStr, ok := element.(string)
		if !ok {
			return false, false
		}
		return true, strings.Contains(listValue.String(), elementStr)
	}

	if listValue.Kind() == reflect.Map {
		mapKeys := listValue.MapKeys()
		for i := 0; i < len(mapKeys); i++ {
			if cmp.Equal(mapKeys[i].Interface(), element) {
				return true, true
			}
		}
		return true, false
	}

	if listValue.Kind() == reflect.Slice || listValue.Kind() == reflect.Array {
		for i := 0; i < listValue.Len(); i++ {
			if cmp.Equal(listValue.Index(i).Interface(), element) {
				return true, true
			}
		}
		return true, false
	}
	return false, false
}

func Equal(t *testing.T, expected, actual any, msgAndArgs ...any) bool {
	t.Helper()
	if !cmp.Equal(expected, actual) {
		t.Errorf("Not equal:\nexpected: %#v\nactual  : %#v\ndiff:\n%s%s", expected, actual, cmp.Diff(expected, actual), formatMsg(msgAndArgs))
		return false
	}
	return true
}

func NotEqual(t *testing.T, expected, actual any, msgAndArgs ...any) bool {
	t.Helper()
	if cmp.Equal(expected, actual) {
		t.Errorf("Should not be: %#v%s", actual, formatMsg(msgAndArgs))
		return false
	}
	return true
}

func Less(t *testing.T, e1, e2 any, msgAndArgs ...any) bool {
	t.Helper()
	v1 := reflect.ValueOf(e1)
	v2 := reflect.ValueOf(e2)
	isLess := false
	switch v1.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		isLess = v1.Int() < v2.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		isLess = v1.Uint() < v2.Uint()
	case reflect.Float32, reflect.Float64:
		isLess = v1.Float() < v2.Float()
	default:
		t.Errorf("Less does not support type %T", e1)
		return false
	}
	if !isLess {
		t.Errorf("Expected %#v to be less than %#v%s", e1, e2, formatMsg(msgAndArgs))
		return false
	}
	return true
}

func Greater(t *testing.T, e1, e2 any, msgAndArgs ...any) bool {
	t.Helper()
	v1 := reflect.ValueOf(e1)
	v2 := reflect.ValueOf(e2)
	isGreater := false
	switch v1.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		isGreater = v1.Int() > v2.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		isGreater = v1.Uint() > v2.Uint()
	case reflect.Float32, reflect.Float64:
		isGreater = v1.Float() > v2.Float()
	default:
		t.Errorf("Greater does not support type %T", e1)
		return false
	}
	if !isGreater {
		t.Errorf("Expected %#v to be greater than %#v%s", e1, e2, formatMsg(msgAndArgs))
		return false
	}
	return true
}

func Len(t *testing.T, object any, length int, msgAndArgs ...any) bool {
	t.Helper()
	v := reflect.ValueOf(object)
	var l int
	var panicked bool
	func() {
		defer func() {
			if e := recover(); e != nil {
				panicked = true
			}
		}()
		l = v.Len()
	}()
	if panicked {
		t.Errorf("Len could not be applied to type %T%s", object, formatMsg(msgAndArgs))
		return false
	}
	if l != length {
		t.Errorf("Expected length %d, got %d%s", length, l, formatMsg(msgAndArgs))
		return false
	}
	return true
}

func Nil(t *testing.T, object any, msgAndArgs ...any) bool {
	t.Helper()
	if object == nil {
		return true
	}
	value := reflect.ValueOf(object)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		if value.IsNil() {
			return true
		}
	}
	t.Errorf("Expected nil, but got: %#v%s", object, formatMsg(msgAndArgs))
	return false
}

func NotNil(t *testing.T, object any, msgAndArgs ...any) bool {
	t.Helper()
	if object == nil {
		t.Errorf("Expected value not to be nil%s", formatMsg(msgAndArgs))
		return false
	}
	value := reflect.ValueOf(object)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		if value.IsNil() {
			t.Errorf("Expected value not to be nil%s", formatMsg(msgAndArgs))
			return false
		}
	}
	return true
}

func Error(t *testing.T, err error, msgAndArgs ...any) bool {
	t.Helper()
	if err == nil {
		t.Errorf("An error is expected but got nil%s", formatMsg(msgAndArgs))
		return false
	}
	return true
}

func NoError(t *testing.T, err error, msgAndArgs ...any) bool {
	t.Helper()
	if err != nil {
		t.Errorf("Received unexpected error:\n%+v%s", err, formatMsg(msgAndArgs))
		return false
	}
	return true
}

func ErrorIs(t *testing.T, err, target error, msgAndArgs ...any) bool {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("Target error should be in err chain:\nexpected: %q\nactual  : %q%s", target, err, formatMsg(msgAndArgs))
		return false
	}
	return true
}

func Empty(t *testing.T, object any, msgAndArgs ...any) bool {
	t.Helper()
	if object == nil {
		return true
	}
	v := reflect.ValueOf(object)
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		if v.Len() != 0 {
			t.Errorf("Expected empty, but has %d item(s)%s", v.Len(), formatMsg(msgAndArgs))
			return false
		}
	case reflect.Ptr:
		if !v.IsNil() {
			t.Errorf("Expected empty pointer, but was not nil%s", formatMsg(msgAndArgs))
			return false
		}
	default:
		zero := reflect.Zero(v.Type()).Interface()
		if !cmp.Equal(object, zero) {
			t.Errorf("Expected empty (zero value), but got: %#v%s", object, formatMsg(msgAndArgs))
			return false
		}
	}
	return true
}

func NotEmpty(t *testing.T, object any, msgAndArgs ...any) bool {
	t.Helper()
	if object == nil {
		t.Errorf("Expected not empty, but got nil%s", formatMsg(msgAndArgs))
		return false
	}
	v := reflect.ValueOf(object)
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		if v.Len() == 0 {
			t.Errorf("Expected not empty, but was%s", formatMsg(msgAndArgs))
			return false
		}
	case reflect.Ptr:
		if v.IsNil() {
			t.Errorf("Expected not empty pointer, but was nil%s", formatMsg(msgAndArgs))
			return false
		}
	default:
		zero := reflect.Zero(v.Type()).Interface()
		if cmp.Equal(object, zero) {
			t.Errorf("Expected not empty (zero value), but got: %#v%s", object, formatMsg(msgAndArgs))
			return false
		}
	}
	return true
}

func True(t *testing.T, value bool, msgAndArgs ...any) bool {
	t.Helper()
	if !value {
		t.Errorf("Expected true, got false%s", formatMsg(msgAndArgs))
		return false
	}
	return true
}

func False(t *testing.T, value bool, msgAndArgs ...any) bool {
	t.Helper()
	if value {
		t.Errorf("Expected false, got true%s", formatMsg(msgAndArgs))
		return false
	}
	return true
}
