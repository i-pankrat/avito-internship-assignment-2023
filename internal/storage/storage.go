package storage

import "errors"

var (
	ErrSegmentExists                = errors.New("segment exists")
	ErrSegmentDoesNotExist          = errors.New("segment does not exists")
	ErrUserHasAlreadyAddedToSegment = errors.New("user has already been added to the segment")
)
