package ecode

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	_codes = map[string]struct{}{}
)

type ECode string

// New create a ecode
func New(e int) ECode {
	if e <= 0 {
		panic("ecode must > 0")
	}

	return add(strconv.Itoa(e))
}

func add(e string) ECode {
	if _, ok := _codes[e]; ok {
		panic(fmt.Sprintf("ecode: %s already exist", e))
	}
	_codes[e] = struct{}{}
	return Code(e)
}

func (e ECode) Ok() bool {
	return string(e) == string(EcodeOk)
}

func (e ECode) Fail() bool {
	return !e.Ok()
}

func (e ECode) Error() string {
	return string(e)
}

func (e ECode) String() string {
	return string(e)
}

// Message return error message
func (e ECode) Message() string {
	if msg, ok := messages[e.Int()]; ok {
		return msg
	}

	return e.Error()
}

// Code from string to ecode
func Code(e string) ECode { return ECode(e) }

// Int from ecode to int
func (e ECode) Int() int {
	i, _ := strconv.Atoi(e.String())
	return i
}

func (e ECode) Is(err error) bool {
	return errors.Is(e, err)
}
