package net

import "errors"

var (
	ErrorServerShutDown         = errors.New("server shut down")
	ErrorConnClosed             = errors.New("conn closed")
	ErrorUnsupportedTCPProtocol = errors.New("ErrorUnsupportedTCPProtocol")
	ErrorTooFewParameters       = errors.New("too few parameters")
)

var (
	ErrorCodeCLengthFieldTooShort = errors.New("codeC lengthField too short")
	ErrorBytesLengthTooShort      = errors.New("bytes length too short")
	ErrorInvalidLengthFieldLength = errors.New("invalid  lengthFieldLength")
)
