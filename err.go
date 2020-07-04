package nnet

import (
	"errors"
)

var (
	ErrExistsAlready          = errors.New("item already exists")
	ErrNotExists              = errors.New("item not exists")
	ErrConnectionReject       = errors.New("connection rejected by logic")
	ErrConnClosing            = errors.New("use of closed network connection")
	ErrWriteBlocking          = errors.New("write packet was blocking")
	ErrReadBlocking           = errors.New("read packet was blocking")
	ErrEmptySlice             = errors.New("the slice is empty")
	ErrSliceOutOfRange        = errors.New("the slice is out of range")
	ErrBufferSizeInsufficient = errors.New("buffer size is too small")
	ErrInterface 						 = errors.New("interface convertion failed")
)
