package types

import "errors"

var (
	ErrInternalError = errors.New("internal error")

	ErrInsufficientResource = errors.New("resource is insufficient")

	ErrContainerNotExist = errors.New("container not exist")
)
