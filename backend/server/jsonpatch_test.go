package server

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tfabritius/plainpage/model"
)

// Test structs for jsonpatch testing
type simpleStruct struct {
	Name  string `json:"name" patch:"allow"`
	Value int    `json:"value" patch:"allow"`
}

type nestedStruct struct {
	Title  string       `json:"title" patch:"allow"`
	Nested simpleStruct `json:"nested" patch:"allow"`
}

type pointerStruct struct {
	Title  string        `json:"title" patch:"allow"`
	Nested *simpleStruct `json:"nested" patch:"allow"`
}

type sliceStruct struct {
	Tags []string `json:"tags" patch:"allow"`
}

type protectedStruct struct {
	Public    string `json:"public" patch:"allow"`
	Protected string `json:"-"`
}

type deeplyNestedStruct struct {
	Level1 *struct {
		Level2 *struct {
			Value string `json:"value" patch:"allow"`
		} `json:"level2" patch:"allow"`
	} `json:"level1" patch:"allow"`
}

// Helper to create json.RawMessage pointer
func rawJSON(s string) *json.RawMessage {
	r := json.RawMessage(s)
	return &r
}

func TestApplyJSONPatch_SimpleString(t *testing.T) {
	target := simpleStruct{Name: "original", Value: 42}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/name", Value: rawJSON(`"updated"`)},
	})

	require.NoError(t, err)
	assert.Equal(t, "updated", target.Name)
	assert.Equal(t, 42, target.Value) // Should be unchanged
}

func TestApplyJSONPatch_SimpleInt(t *testing.T) {
	target := simpleStruct{Name: "test", Value: 10}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/value", Value: rawJSON(`99`)},
	})

	require.NoError(t, err)
	assert.Equal(t, 99, target.Value)
}

func TestApplyJSONPatch_MultipleOperations(t *testing.T) {
	target := simpleStruct{Name: "original", Value: 10}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/name", Value: rawJSON(`"updated"`)},
		{Op: "replace", Path: "/value", Value: rawJSON(`20`)},
	})

	require.NoError(t, err)
	assert.Equal(t, "updated", target.Name)
	assert.Equal(t, 20, target.Value)
}

func TestApplyJSONPatch_NestedStruct(t *testing.T) {
	target := nestedStruct{
		Title:  "title",
		Nested: simpleStruct{Name: "nested", Value: 5},
	}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/nested/name", Value: rawJSON(`"updated nested"`)},
	})

	require.NoError(t, err)
	assert.Equal(t, "updated nested", target.Nested.Name)
	assert.Equal(t, 5, target.Nested.Value) // Should be unchanged
}

func TestApplyJSONPatch_ReplaceEntireNestedStruct(t *testing.T) {
	target := nestedStruct{
		Title:  "title",
		Nested: simpleStruct{Name: "original", Value: 5},
	}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/nested", Value: rawJSON(`{"name": "new", "value": 100}`)},
	})

	require.NoError(t, err)
	assert.Equal(t, "new", target.Nested.Name)
	assert.Equal(t, 100, target.Nested.Value)
}

func TestApplyJSONPatch_PointerField_Initialize(t *testing.T) {
	target := pointerStruct{Title: "title", Nested: nil}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/nested/name", Value: rawJSON(`"initialized"`)},
	})

	require.NoError(t, err)
	require.NotNil(t, target.Nested)
	assert.Equal(t, "initialized", target.Nested.Name)
}

func TestApplyJSONPatch_PointerField_Update(t *testing.T) {
	target := pointerStruct{
		Title:  "title",
		Nested: &simpleStruct{Name: "original", Value: 10},
	}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/nested/value", Value: rawJSON(`20`)},
	})

	require.NoError(t, err)
	assert.Equal(t, 20, target.Nested.Value)
	assert.Equal(t, "original", target.Nested.Name) // unchanged
}

func TestApplyJSONPatch_PointerField_SetToNull(t *testing.T) {
	target := pointerStruct{
		Title:  "title",
		Nested: &simpleStruct{Name: "original", Value: 10},
	}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/nested", Value: rawJSON(`null`)},
	})

	require.NoError(t, err)
	assert.Nil(t, target.Nested)
}

func TestApplyJSONPatch_SliceField(t *testing.T) {
	target := sliceStruct{Tags: []string{"a", "b"}}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/tags", Value: rawJSON(`["x", "y", "z"]`)},
	})

	require.NoError(t, err)
	assert.Equal(t, []string{"x", "y", "z"}, target.Tags)
}

func TestApplyJSONPatch_SliceField_SetToNull(t *testing.T) {
	target := sliceStruct{Tags: []string{"a", "b"}}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/tags", Value: rawJSON(`null`)},
	})

	require.NoError(t, err)
	assert.Nil(t, target.Tags)
}

func TestApplyJSONPatch_SliceField_SetToEmpty(t *testing.T) {
	target := sliceStruct{Tags: []string{"a", "b"}}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/tags", Value: rawJSON(`[]`)},
	})

	require.NoError(t, err)
	assert.Empty(t, target.Tags)
	assert.NotNil(t, target.Tags)
}

func TestApplyJSONPatch_DeeplyNested_InitializeAlongPath(t *testing.T) {
	target := deeplyNestedStruct{}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/level1/level2/value", Value: rawJSON(`"deep"`)},
	})

	require.NoError(t, err)
	require.NotNil(t, target.Level1)
	require.NotNil(t, target.Level1.Level2)
	assert.Equal(t, "deep", target.Level1.Level2.Value)
}

func TestApplyJSONPatch_EmptyOperations(t *testing.T) {
	target := simpleStruct{Name: "original", Value: 42}

	err := ApplyJSONPatch(&target, []model.PatchOperation{})

	require.NoError(t, err)
	assert.Equal(t, "original", target.Name)
	assert.Equal(t, 42, target.Value)
}

// Error cases

func TestApplyJSONPatch_Error_UnsupportedOperation(t *testing.T) {
	target := simpleStruct{}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "add", Path: "/name", Value: rawJSON(`"test"`)},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "operation add not supported")
}

func TestApplyJSONPatch_Error_UnsupportedOperation_Remove(t *testing.T) {
	target := simpleStruct{}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "remove", Path: "/name"},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "operation remove not supported")
}

func TestApplyJSONPatch_Error_PathNotStartingWithSlash(t *testing.T) {
	target := simpleStruct{}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "name", Value: rawJSON(`"test"`)},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "path must start with /")
}

func TestApplyJSONPatch_Error_EmptyPath(t *testing.T) {
	target := simpleStruct{}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/", Value: rawJSON(`"test"`)},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty path")
}

func TestApplyJSONPatch_Error_UnknownField(t *testing.T) {
	target := simpleStruct{}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/unknown", Value: rawJSON(`"test"`)},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "path not supported")
}

func TestApplyJSONPatch_Error_UnknownNestedField(t *testing.T) {
	target := nestedStruct{}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/nested/unknown", Value: rawJSON(`"test"`)},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "path not supported")
}

func TestApplyJSONPatch_Error_ProtectedField(t *testing.T) {
	target := protectedStruct{Public: "public", Protected: "protected"}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/Protected", Value: rawJSON(`"hacked"`)},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "path not supported")
	assert.Equal(t, "protected", target.Protected) // unchanged
}

func TestApplyJSONPatch_Error_NullOnNonNullableString(t *testing.T) {
	target := simpleStruct{Name: "test"}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/name", Value: rawJSON(`null`)},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot set null on non-nullable field")
}

func TestApplyJSONPatch_Error_NullOnNonNullableInt(t *testing.T) {
	target := simpleStruct{Value: 42}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/value", Value: rawJSON(`null`)},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot set null on non-nullable field")
}

func TestApplyJSONPatch_Error_NullOnNonNullableNestedString(t *testing.T) {
	target := nestedStruct{
		Nested: simpleStruct{Name: "test"},
	}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/nested/name", Value: rawJSON(`null`)},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot set null on non-nullable field")
}

func TestApplyJSONPatch_Error_InvalidJSON(t *testing.T) {
	target := simpleStruct{}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/name", Value: rawJSON(`{invalid}`)},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid value")
}

func TestApplyJSONPatch_Error_TypeMismatch(t *testing.T) {
	target := simpleStruct{}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/value", Value: rawJSON(`"not a number"`)},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid value")
}

func TestApplyJSONPatch_Error_NavigateIntoNonStruct(t *testing.T) {
	target := simpleStruct{Name: "test"}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/name/subfield", Value: rawJSON(`"test"`)},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot navigate into non-struct")
}

func TestApplyJSONPatch_StopsOnFirstError(t *testing.T) {
	target := simpleStruct{Name: "original", Value: 10}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/name", Value: rawJSON(`"updated"`)},
		{Op: "add", Path: "/value", Value: rawJSON(`20`)}, // This should fail
		{Op: "replace", Path: "/value", Value: rawJSON(`30`)},
	})

	require.Error(t, err)
	assert.Equal(t, "updated", target.Name) // First op succeeded
	assert.Equal(t, 10, target.Value)       // Second op failed, value unchanged
}

// Test with json tag options (omitempty, etc.)

type tagOptionsStruct struct {
	Name     string `json:"name,omitempty" patch:"allow"`
	Required string `json:"required" patch:"allow"`
}

func TestApplyJSONPatch_JsonTagWithOptions(t *testing.T) {
	target := tagOptionsStruct{Name: "original", Required: "req"}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/name", Value: rawJSON(`"updated"`)},
	})

	require.NoError(t, err)
	assert.Equal(t, "updated", target.Name)
}

// Test that fields without patch:"allow" tag are rejected
type patchTagTestStruct struct {
	Allowed    string `json:"allowed" patch:"allow"`
	NotAllowed string `json:"notAllowed"` // no patch tag - should be rejected
}

func TestApplyJSONPatch_Error_FieldWithoutPatchTag(t *testing.T) {
	target := patchTagTestStruct{Allowed: "a", NotAllowed: "b"}

	// Patching a field without patch:"allow" should fail
	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/notAllowed", Value: rawJSON(`"hacked"`)},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "path not supported")
	assert.Equal(t, "b", target.NotAllowed) // unchanged

	// But patching a field with patch:"allow" should work
	err = ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/allowed", Value: rawJSON(`"updated"`)},
	})

	require.NoError(t, err)
	assert.Equal(t, "updated", target.Allowed)
}

// Test that error messages include path

func TestApplyJSONPatch_ErrorIncludesPath(t *testing.T) {
	target := simpleStruct{}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/unknown", Value: rawJSON(`"test"`)},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "/unknown")
}

func TestApplyJSONPatch_ErrorIncludesNestedPath(t *testing.T) {
	target := nestedStruct{}

	err := ApplyJSONPatch(&target, []model.PatchOperation{
		{Op: "replace", Path: "/nested/unknown", Value: rawJSON(`"test"`)},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "/nested/unknown")
}
