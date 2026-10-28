package require

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

func Contains(t *testing.T, s, contains any, msgAndArgs ...any) {
	t.Helper()
	ok, found := includeElement(s, contains)
	if !ok {
		t.Fatalf("Contains could not be applied to type %T%s", s, formatMsg(msgAndArgs))
	}
	if !found {
		t.Fatalf("%#v does not contain %#v%s", s, contains, formatMsg(msgAndArgs))
	}
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

func Equal(t *testing.T, expected, actual any, msgAndArgs ...any) {
	t.Helper()
	if !cmp.Equal(expected, actual) {
		t.Fatalf("Not equal:\nexpected: %#v\nactual  : %#v\ndiff:\n%s%s", expected, actual, cmp.Diff(expected, actual), formatMsg(msgAndArgs))
	}
}

func NotEqual(t *testing.T, expected, actual any, msgAndArgs ...any) {
	t.Helper()
	if cmp.Equal(expected, actual) {
		t.Fatalf("Should not be: %#v%s", actual, formatMsg(msgAndArgs))
	}
}

func Less(t *testing.T, e1, e2 any, msgAndArgs ...any) {
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
		t.Fatalf("Less does not support type %T", e1)
	}
	if !isLess {
		t.Fatalf("Expected %#v to be less than %#v%s", e1, e2, formatMsg(msgAndArgs))
	}
}

func Greater(t *testing.T, e1, e2 any, msgAndArgs ...any) {
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
		t.Fatalf("Greater does not support type %T", e1)
	}
	if !isGreater {
		t.Fatalf("Expected %#v to be greater than %#v%s", e1, e2, formatMsg(msgAndArgs))
	}
}

func Len(t *testing.T, object any, length int, msgAndArgs ...any) {
	t.Helper()
	v := reflect.ValueOf(object)
	defer func() {
		if e := recover(); e != nil {
			t.Fatalf("Len could not be applied to type %T%s", object, formatMsg(msgAndArgs))
		}
	}()
	l := v.Len()
	if l != length {
		t.Fatalf("Expected length %d, got %d%s", length, l, formatMsg(msgAndArgs))
	}
}

func Nil(t *testing.T, object any, msgAndArgs ...any) {
	t.Helper()
	if object == nil {
		return
	}
	value := reflect.ValueOf(object)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		if value.IsNil() {
			return
		}
	}
	t.Fatalf("Expected nil, but got: %#v%s", object, formatMsg(msgAndArgs))
}

func NotNil(t *testing.T, object any, msgAndArgs ...any) {
	t.Helper()
	if object == nil {
		t.Fatalf("Expected value not to be nil%s", formatMsg(msgAndArgs))
	}
	value := reflect.ValueOf(object)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		if value.IsNil() {
			t.Fatalf("Expected value not to be nil%s", formatMsg(msgAndArgs))
		}
	}
}

func Error(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	if err == nil {
		t.Fatalf("An error is expected but got nil%s", formatMsg(msgAndArgs))
	}
}

func NoError(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	if err != nil {
		t.Fatalf("Received unexpected error:\n%+v%s", err, formatMsg(msgAndArgs))
	}
}

func ErrorIs(t *testing.T, err, target error, msgAndArgs ...any) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("Target error should be in err chain:\nexpected: %q\nactual  : %q%s", target, err, formatMsg(msgAndArgs))
	}
}

func Empty(t *testing.T, object any, msgAndArgs ...any) {
	t.Helper()
	if object == nil {
		return
	}
	v := reflect.ValueOf(object)
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		if v.Len() != 0 {
			t.Fatalf("Expected empty, but has %d item(s)%s", v.Len(), formatMsg(msgAndArgs))
		}
	case reflect.Ptr:
		if !v.IsNil() {
			t.Fatalf("Expected empty pointer, but was not nil%s", formatMsg(msgAndArgs))
		}
	default:
		zero := reflect.Zero(v.Type()).Interface()
		if !cmp.Equal(object, zero) {
			t.Fatalf("Expected empty (zero value), but got: %#v%s", object, formatMsg(msgAndArgs))
		}
	}
}

func NotEmpty(t *testing.T, object any, msgAndArgs ...any) {
	t.Helper()
	if object == nil {
		t.Fatalf("Expected not empty, but got nil%s", formatMsg(msgAndArgs))
	}
	v := reflect.ValueOf(object)
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		if v.Len() == 0 {
			t.Fatalf("Expected not empty, but was%s", formatMsg(msgAndArgs))
		}
	case reflect.Ptr:
		if v.IsNil() {
			t.Fatalf("Expected not empty pointer, but was nil%s", formatMsg(msgAndArgs))
		}
	default:
		zero := reflect.Zero(v.Type()).Interface()
		if cmp.Equal(object, zero) {
			t.Fatalf("Expected not empty (zero value), but got: %#v%s", object, formatMsg(msgAndArgs))
		}
	}
}

func True(t *testing.T, value bool, msgAndArgs ...any) {
	t.Helper()
	if !value {
		t.Fatalf("Expected true, got false%s", formatMsg(msgAndArgs))
	}
}

func False(t *testing.T, value bool, msgAndArgs ...any) {
	t.Helper()
	if value {
		t.Fatalf("Expected false, got true%s", formatMsg(msgAndArgs))
	}
}
