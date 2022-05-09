package net

import "errors"

var (
	ErrorServerShutDown = errors.New("server shut down")
)

var (
	ErrorCodeCLengthFieldTooShort = errors.New("codeC lengthField too short")
	ErrorBytesLengthTooShort      = errors.New("bytes length too short")
	ErrorInvalidLengthFieldLength = errors.New("invalid  lengthFieldLength")
)
