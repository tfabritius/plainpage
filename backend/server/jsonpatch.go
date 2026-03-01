package server

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/tfabritius/plainpage/model"
)

// ApplyJSONPatch applies RFC 6902 patch operations to a struct using reflection.
// It uses json tags to map paths like "/retention/trash/maxAgeDays" to struct fields.
// Only the "replace" operation is supported.
// Fields with json:"-" tag are protected and cannot be patched.
// Only fields with patch:"allow" tag can be patched.
// Nil values are allowed and will set the field to its JSON null value.
func ApplyJSONPatch[T any](target *T, operations []model.PatchOperation) error {
	for _, op := range operations {
		if op.Op != "replace" {
			return fmt.Errorf("operation %s not supported", op.Op)
		}

		// Use "null" JSON value if op.Value is nil
		var rawValue json.RawMessage
		if op.Value == nil {
			rawValue = json.RawMessage("null")
		} else {
			rawValue = *op.Value
		}

		if err := setFieldByJSONPath(target, op.Path, rawValue); err != nil {
			return fmt.Errorf("error at %s: %w", op.Path, err)
		}
	}
	return nil
}

// setFieldByJSONPath navigates the struct using json tags and sets the value.
// It automatically initializes nil pointers along the path.
func setFieldByJSONPath(target any, path string, value json.RawMessage) error {
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path must start with /")
	}

	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) == 0 || (len(parts) == 1 && parts[0] == "") {
		return fmt.Errorf("empty path")
	}

	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Navigate to the nested field
	for i, part := range parts {
		if v.Kind() != reflect.Struct {
			return fmt.Errorf("cannot navigate into non-struct at %s", part)
		}

		field, found := findFieldByJSONTag(v.Type(), part)
		if !found {
			return fmt.Errorf("path not supported")
		}

		v = v.FieldByIndex(field.Index)

		// Last part - set the value
		if i == len(parts)-1 {
			if !v.CanSet() {
				return fmt.Errorf("cannot set field")
			}

			// Check if trying to set null on a non-nullable field
			if string(value) == "null" {
				kind := v.Type().Kind()
				// Allow null only for pointer, slice, map, interface types and custom types that implement json.Unmarshaler
				if kind != reflect.Ptr && kind != reflect.Slice && kind != reflect.Map && kind != reflect.Interface {
					// Check if the type implements json.Unmarshaler
					if !reflect.PointerTo(v.Type()).Implements(reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()) {
						return fmt.Errorf("cannot set null on non-nullable field")
					}
				}
			}

			// Create new value and unmarshal
			newVal := reflect.New(v.Type())
			if err := json.Unmarshal(value, newVal.Interface()); err != nil {
				return fmt.Errorf("invalid value: %w", err)
			}
			v.Set(newVal.Elem())
			return nil
		}

		// Navigate into pointer if needed, initializing nil pointers
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				// Initialize the nil pointer with a new value
				if !v.CanSet() {
					return fmt.Errorf("cannot initialize nil pointer at %s", part)
				}
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
	}

	return fmt.Errorf("path not found")
}

// findFieldByJSONTag finds a struct field by its json tag name.
// Only fields with patch:"allow" tag can be patched.
func findFieldByJSONTag(t reflect.Type, name string) (reflect.StructField, bool) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")

		// Skip fields with json:"-" (protected fields)
		if jsonTag == "-" {
			continue
		}

		// Extract the json field name (before any comma)
		jsonName := strings.Split(jsonTag, ",")[0]

		if jsonName == name {
			// Check if field has patch:"allow" tag
			patchTag := field.Tag.Get("patch")
			if patchTag != "allow" {
				// Field exists but patching is not allowed - return not found
				// to give the same error message as non-existent fields
				return reflect.StructField{}, false
			}
			return field, true
		}
	}
	return reflect.StructField{}, false
}
