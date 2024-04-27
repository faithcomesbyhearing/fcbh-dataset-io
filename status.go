package dataset

import (
	"strconv"
	"strings"
)

type Status struct {
	IsErr   bool   `json:"is_error"`
	Message string `json:"message"`
	Status  int    `json:"status"`
	Err     string `json:"error,omitempty"`
	Request string `json:"request"`
	Trace   string `json:"trace,omitempty"`
}

// Status implements the Error interface
func (e *Status) Error() string {
	return e.String()
}

// Status implements the Stringer interface
// Using fmt package here caused stack overflow
func (e *Status) String() string {
	var result = make([]string, 0)
	result = append(result, `{ "is_error": `+strconv.FormatBool(e.IsErr))
	result = append(result, ` "message": "`+e.Message+`"`)
	result = append(result, ` "status": `+strconv.Itoa(e.Status))
	result = append(result, ` "error": "`+e.Err+`" }`)
	return strings.Join(result, ",")
}
