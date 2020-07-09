package apierrors

import "errors"

// A list of error messages for Search API
var (
	ErrEmptySearchTerm        = errors.New("empty search term")
	ErrIndexNotFound          = errors.New("search index not found")
	ErrInternalServer         = errors.New("internal server error")
	ErrMarshallingQuery       = errors.New("failed to marshal query to bytes for request body to send to elastic")
	ErrParsingQueryParameters = errors.New("failed to parse query parameters, values must be an integer")
	ErrUnmarshallingJSON      = errors.New("failed to parse json body")
	ErrUnexpectedStatusCode   = errors.New("unexpected status code from elastic api")

	NotFoundMap = map[error]bool{}

	BadRequestMap = map[error]bool{
		ErrEmptySearchTerm:        true,
		ErrParsingQueryParameters: true,
	}
)