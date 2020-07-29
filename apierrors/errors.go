package apierrors

import "errors"

// A list of error messages for Search API
var (
	ErrBadSearchQuery          = errors.New("bad query sent to elasticsearch index")
	ErrEmptySearchTerm         = errors.New("empty search term")
	ErrIndexNotFound           = errors.New("search index not found")
	ErrInternalServer          = errors.New("internal server error")
	ErrMarshallingQuery        = errors.New("failed to marshal query to bytes for request body to send to elastic")
	ErrParsingQueryParameters  = errors.New("failed to parse query parameters, values must be an integer")
	ErrTooManyDimensionFilters = errors.New("Too many dimension filters, limited to a maximum of 10")
	ErrTooManyTopicFilters     = errors.New("Too many topic filters, limited to a maximum of 10")
	ErrTopicNotFound           = errors.New("Topic not found")
	ErrUnmarshallingJSON       = errors.New("failed to parse json body")
	ErrUnexpectedStatusCode    = errors.New("unexpected status code from elastic api")

	NotFoundMap = map[error]bool{
		ErrTopicNotFound: true,
	}

	BadRequestMap = map[error]bool{
		ErrEmptySearchTerm:         true,
		ErrParsingQueryParameters:  true,
		ErrTooManyDimensionFilters: true,
		ErrTooManyTopicFilters:     true,
	}
)
