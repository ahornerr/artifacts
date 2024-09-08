package httperror

import (
	"encoding/json"
	"errors"
	"fmt"
)

type rawHttpError struct {
	Error struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

type HTTPError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Body    string
}

func NewHTTPError(statusCode int, body []byte) HTTPError {
	var httpError HTTPError

	var raw rawHttpError
	err := json.Unmarshal(body, &raw)
	if err != nil {
		switch statusCode {
		case 502:
			httpError.Body = "502 Bad Gateway"
		case 503:
			httpError.Body = "503 Service Unavailable"
		default:
			httpError.Body = string(body)
		}
		httpError.Message = err.Error()
	} else {
		httpError.Body = string(body)
		httpError.Code = raw.Error.Code
		httpError.Message = raw.Error.Message
	}

	return httpError
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("[%d] %s - %s", e.Code, e.Message, e.Body)
}

func ErrIsBankItemNotFound(err error) bool {
	var httpError HTTPError
	if !errors.As(err, &httpError) {
		return false
	}

	if httpError.Code != 404 {
		return false
	}

	return httpError.Message == "Item not found."
}

func ErrIsBankInsufficientQuantity(err error) bool {
	var httpError HTTPError
	if !errors.As(err, &httpError) {
		return false
	}

	if httpError.Code != 478 {
		return false
	}

	return httpError.Message == "Missing item or insufficient quantity."
}
