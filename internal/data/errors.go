package data

import (
	"errors"
)

var ErrRecordNotFound = errors.New("record not found")
var QueryFail = errors.New("error in SQL query")
