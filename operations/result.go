package operations

import "github.com/beyondstorage/go-storage/v4/types"

// The idea of XXXResult is borrowed from rust Result<Object>
// An empty result in rust would be Result<()>

// EmptyResult is the result for empty return.
// Only Error field will be valid.
// Used in process like copy, which does not return an Object.
type EmptyResult struct {
	Error error
}

// ObjectResult is the result for Object.
// Only one of Object or Error will be valid.
// We need to check Error before use Object.
type ObjectResult struct {
	Object *types.Object
	Error  error
}

// PartResult is the result for Part.
// Only one of Part or Error will be valid.
// We need to check Error before use Part.
type PartResult struct {
	Part  *types.Part
	Error error
}
