package coprhd

import (
	"strings"
)

const (
	ErrCodeOK               = 0
	ErrCodeInvalidParam     = 1008
	ErrCodeCreateNotAllowed = 1054
)

func (err ApiError) IsOK() bool {
	return err.Code == ErrCodeOK
}

func (err ApiError) IsDup() bool {
	return strings.Contains(err.Details, "already exists")
}
