package musicextractors

import "errors"

var (
	ErrNoURLFound           = errors.New("no URL found in text")
	ErrMultipleResult       = errors.New("multiple results found in string")
	ErrInvalidRegexCompiled = errors.New("invalid regex condition")
)

var (
	ErrNoTitleFound  = errors.New("no title found in page")
	ErrRequestFailed = errors.New("failed to fetch URL")
)
