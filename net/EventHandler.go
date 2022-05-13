package net

type EventHandler interface {
	PreWrite(c Conn, b []byte)
	AfterWrite(c Conn)
	React(msg []byte, c Conn) ([]byte, error)
}

type DefaultHandler int

func NewDefaultHandler() EventHandler {
	var i DefaultHandler
	return &i
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
