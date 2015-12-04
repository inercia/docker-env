package config

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// Variables must be like {{.VAR}}
var varRegexp = regexp.MustCompile(`\$\(.*?\)`)

// check if the string has a variable
func hasVar(s string, v string) bool {
	return strings.Contains(s, "$(#)")
}

// check if the string has (at least) one variable like {{.VAR}}
func hasVars(s string) bool {
	return varRegexp.MatchString(s)
}

// replace variables like {{.VAR}}
func replaceVars(s string, vars map[string]string) (string, error) {
	unknown := []string{}
	repl := func(k string) string {
		kk := k[2 : len(k)-1] // take out the ".{{" and the "}}"
		if v, found := vars[kk]; found {
			return v
		} else {
			unknown = append(unknown, kk)
			return k
		}
	}

	replaced := varRegexp.ReplaceAllStringFunc(s, repl)
	if len(unknown) > 0 {
		return replaced, fmt.Errorf("unknown variable(s): %v", unknown)
	}
	return replaced, nil
}

func ReplaceAllStringMap(obj interface{}, replacements map[string]string) interface{} {
	return ReplaceAllStringFunc(obj, func(in string) string {
		replaced, err := replaceVars(in, replacements)
		if err != nil {
			return in
		}
		return replaced
	})
}

func ReplaceAllStringFunc(obj interface{}, replacer func(string) string) interface{} {
	// Wrap the original in a reflect.Value
	original := reflect.ValueOf(obj)

	copy := reflect.New(original.Type()).Elem()
	replaceRecursive(copy, original, replacer)

	// Remove the reflection wrapper
	return copy.Interface()
}

func replaceRecursive(copy, original reflect.Value, replacer func(string) string) {
	switch original.Kind() {
	// The first cases handle nested structures and translate them recursively

	// If it is a pointer we need to unwrap and call once again
	case reflect.Ptr:
		// To get the actual value of the original we have to call Elem()
		// At the same time this unwraps the pointer so we don't end up in
		// an infinite recursion
		originalValue := original.Elem()
		// Check if the pointer is nil
		if !originalValue.IsValid() {
			return
		}
		// Allocate a new object and set the pointer to it
		copy.Set(reflect.New(originalValue.Type()))
		// Unwrap the newly created pointer
		replaceRecursive(copy.Elem(), originalValue, replacer)

	// If it is an interface (which is very similar to a pointer), do basically the
	// same as for the pointer. Though a pointer is not the same as an interface so
	// note that we have to call Elem() after creating a new object because otherwise
	// we would end up with an actual pointer
	case reflect.Interface:
		// Get rid of the wrapping interface
		originalValue := original.Elem()
		// Create a new object. Now new gives us a pointer, but we want the value it
		// points to, so we have to call Elem() to unwrap it
		copyValue := reflect.New(originalValue.Type()).Elem()
		replaceRecursive(copyValue, originalValue, replacer)
		copy.Set(copyValue)

	// If it is a struct we translate each field
	case reflect.Struct:
		for i := 0; i < original.NumField(); i += 1 {
			if original.Field(i).CanInterface() {
				replaceRecursive(copy.Field(i), original.Field(i), replacer)
			}
		}

	// If it is a slice we create a new slice and translate each element
	case reflect.Slice:
		copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i += 1 {
			replaceRecursive(copy.Index(i), original.Index(i), replacer)
		}

	// If it is a map we create a new map and translate each value
	case reflect.Map:
		if original.CanInterface() {
			copy.Set(reflect.MakeMap(original.Type()))
			for _, key := range original.MapKeys() {
				originalValue := original.MapIndex(key)
				// New gives us a pointer, but again we want the value
				copyValue := reflect.New(originalValue.Type()).Elem()
				replaceRecursive(copyValue, originalValue, replacer)
				copy.SetMapIndex(key, copyValue)
			}
		}

	// Otherwise we cannot traverse anywhere so this finishes the the recursion

	// If it is a string translate it (yay finally we're doing what we came for)
	case reflect.String:
		if original.CanInterface() {
			translatedString := replacer(original.Interface().(string))
			copy.SetString(translatedString)
		}

	// And everything else will simply be taken from the original
	default:
		copy.Set(original)
	}

}
