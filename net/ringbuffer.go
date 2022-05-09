package net

import "github.com/valyala/bytebufferpool"

type ringBuffer struct {
	read  int
	write int
	buf   []byte
	bytebufferpool.Pool
}

//Size()
//Len()
//IsEmpty()
//Read(p []byte)(n int,err error)
//Write(p []byte)(n int,err error)
//PeekN()[]byte
//PeekAll()[]byte
//ShiftN(n int)
//Reset()
//ReadWithBytes(p []bytes)([]bytes)
