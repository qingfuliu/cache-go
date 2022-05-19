package cache_go

import "errors"

var (
	KeyDoesNotExists   = errors.New("no such key")
	GroupDoesNotExists = errors.New("no such group")
	ErrorServerBusy    = errors.New("server busy")
)
