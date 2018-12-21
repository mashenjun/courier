package com

type ServiceError struct {
	ErrorCode int `json:"error_code"`
	StatusCode int `json:"-"`
	Message string `json:"message"`
}

func (se *ServiceError) Error() string {
	return se.Message
}

var (
	ParameterError = &ServiceError{ErrorCode: 400000 , StatusCode: 400, Message: "invalid parameters"}

	InternalError = &ServiceError{ ErrorCode: 500000, StatusCode: 500, Message: "sorry, we made a mistake"}
	)

