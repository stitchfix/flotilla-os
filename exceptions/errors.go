package exceptions

//
// MalformedInput describes malformed or otherwise incorrect input
//
type MalformedInput struct {
	ErrorString string
}

func (e MalformedInput) Error() string {
	return e.ErrorString
}

//
// ConflictingResource describes a conflict case:
// eg. definition already exists, reserved fields
//
type ConflictingResource struct {
	ErrorString string
}

func (e ConflictingResource) Error() string {
	return e.ErrorString
}

//
// ResourceMissing describes case where a resource does not exist
// eg. missing definition or run or no image found
//
type MissingResource struct {
	ErrorString string
}

func (e MissingResource) Error() string {
	return e.ErrorString
}
