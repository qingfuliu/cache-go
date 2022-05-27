package net

type EventHandler interface {
	OnEstablish(c Conn) error
	PreWrite(c Conn, b []byte)
	AfterWrite(c Conn)
	React(msg []byte, c Conn) ([]byte, error)
}

type DefaultHandler int

func NewDefaultHandler() EventHandler {
	var i DefaultHandler
	return &i
}

func (e *DefaultHandler) OnEstablish(c Conn) error {
	return nil
}

func (e *DefaultHandler) PreWrite(c Conn, b []byte) {
	return
}
func (e *DefaultHandler) AfterWrite(c Conn) {
	return
}
func (e *DefaultHandler) React(msg []byte, c Conn) ([]byte, error) {
	return nil, nil
}
