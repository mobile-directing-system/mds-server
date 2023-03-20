package wsutil

import "github.com/lefinal/meh"

// TypeError is used or error messages.
const TypeError MessageType = "_error"

// MessageError is the message content for TypeError messages.
type MessageError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
}

// ErrorMessageFromErr genreates a MessageError from the given error which is
// used as content for TypeError messages.
func ErrorMessageFromErr(err error) MessageError {
	e := meh.Cast(err)
	m := MessageError{
		Code:    string(meh.ErrorCode(e)),
		Message: e.Error(),
		Details: e.Details,
	}
	return m
}
