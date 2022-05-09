package byteString

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

type ByteString struct {
	buf []byte
	str string
}

func NewByteString() ByteString {
	return ByteString{}
}

func (bs ByteString) Len() int {
	if bs.buf != nil {
		return len(bs.buf)
	}
	return len(bs.str)
}
func (bs ByteString) At(idx int) byte {
	if bs.buf != nil {
		return bs.buf[idx]
	}
	return bs.str[idx]
}
func (bs ByteString) Slice(begin, end int) ByteString {
	if bs.buf != nil {
		return ByteString{
			buf: bs.buf[begin:end],
		}
	}
	return ByteString{
		str: bs.str[begin:end],
	}
}
func (bs ByteString) SliceBegin(begin int) ByteString {
	if bs.buf != nil {
		return ByteString{
			buf: bs.buf[begin:],
		}
	}
	return ByteString{
		str: bs.str[begin:],
	}
}
func (bs ByteString) SliceEnd(end int) ByteString {
	if bs.buf != nil {
		return ByteString{
			buf: bs.buf[:end],
		}
	}
	return ByteString{
		str: bs.str[:end],
	}
}
func (bs ByteString) String() string {
	if bs.buf != nil {
		return string(bs.buf)
	}
	return bs.str
}

func (bs ByteString) Bytes() []byte {
	if bs.buf != nil {
		return bs.buf
	}
	return []byte(bs.str)
}

func (bs ByteString) Copy(dest []byte) int {
	if bs.buf != nil {
		return copy(dest, bs.buf)
	}
	return copy(dest, bs.str)
}

func (bs ByteString) Equal(b ByteString) bool {
	if b.buf != nil {
		return bs.EqualBytes(b.buf)
	}
	return bs.EqualString(b.str)
}
func (bs ByteString) EqualBytes(b []byte) bool {
	if bs.buf != nil {
		return bytes.Equal(b, bs.buf)
	}

	if len(b) != len(bs.str) {
		return false
	}

	for i := range b {
		if b[i] != bs.str[i] {
			return false
		}
	}
	return true
}

func (bs ByteString) EqualString(s string) bool {
	if bs.buf != nil {

		if len(bs.buf) != len(s) {
			return false
		}

		for i := range s {
			if s[i] != bs.buf[i] {
				return false
			}
		}
		return true
	}

	return s == bs.str
}

func (bs ByteString) Reader() io.ReadSeeker {
	if bs.buf == nil {
		return bytes.NewReader(bs.buf)
	}
	return strings.NewReader(bs.str)
}

func (bs ByteString) ReadAt(buf []byte, offset int64) (n int, err error) {
	if offset < 0 {
		return 0, errors.New("bytes.Reader.ReadAt: negative offset")
	}
	if offset > int64(bs.Len()) {
		return 0, io.EOF
	}

	n = bs.SliceBegin(int(offset)).Copy(buf)

	if n < len(buf) {
		err = io.EOF
	}

	return
}
func (bs ByteString) WriteTo(w io.Writer) (n int64, err error) {
	var m int
	if bs.buf != nil {
		m, err = w.Write(bs.buf)
	} else {
		m, err = io.WriteString(w, bs.str)
	}

	if m < bs.Len() {
		err = io.ErrShortWrite
	}
	n = int64(m)
	return
}
