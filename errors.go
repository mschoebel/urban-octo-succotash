package uos

import (
	"errors"
)

var (
	// ErrorFragmentNotFound is returned if the requested fragment is not available
	ErrorFragmentNotFound = errors.New("fragment not found")
	// ErrorFragmentInvalidRequest is returned if the fragment request parameters are invalid
	ErrorFragmentInvalidRequest = errors.New("invalid fragment request")

	// ErrorFormItemNotFound is returned if the requested item was not found (by 'id')
	ErrorFormItemNotFound = errors.New("form item not found")
	// ErrorFormInvalidRequest is returned if the form request parameters are invalid
	ErrorFormInvalidRequest = errors.New("invalid form request")

	// ErrorTableInvaliRequest is returned if the table request parameters are invalid
	ErrorTableInvalidRequest = errors.New("invalid table request")

	// ErrorInvalidPassword is returned if user authentication credentials are invalid
	ErrorInvalidPassword = errors.New("invalid user credentials")
)
