package shared

import (
	"time"
)

var Interval int64 = 50 //= 100 //8ms

var RequestTimeout = time.Second * 10

var SendCount int64 = 0
var ActiveCount int64 = 0
var DoneCount int64 = 0
var FailedCount int64 = 0

var TestMode bool
var DevMode bool
var LocalMode bool

func VariableWrapper[T any](anyValue T) T {
	return anyValue
}

func VariablePtrWrapper[T any](anyValue T) *T {
	return &anyValue
}

func When[T any](c bool, d1, d2 T) T {
	if c {
		return d1
	} else {
		return d2
	}
}

var Addr string
var Key string
