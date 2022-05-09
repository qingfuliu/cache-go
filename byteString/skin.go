package byteString

import (
	"github.com/golang/protobuf/proto"
	"unsafe"
)

type Skin interface {
	SetString(str string)
	SetBytes(bys []byte)
	SetProto(pro proto.Message)
	SetByteString(byteString ByteString)
	View() ByteString
}

// byteSkin ----------------------byteSkin---------------------- byteSkin//

type byteSkin struct {
	bytes *[]byte
	bs    ByteString
}

func NewByteSkin(bytePointer *[]byte) Skin {
	if bytePointer == nil {
		return nil
	}
	return &byteSkin{
		bytes: bytePointer,
		bs: ByteString{
			buf: []byte{},
		},
	}
}

func (BS *byteSkin) SetString(str string) {
	BS.bs.buf = append(BS.bs.buf[:0], str...)
	*BS.bytes = BS.bs.buf
}
func (BS *byteSkin) SetBytes(bys []byte) {
	BS.bs.buf = BS.bs.buf[:0]
	BS.bs.buf = append(BS.bs.buf, bys...)
	*BS.bytes = BS.bs.buf
}
func (BS *byteSkin) SetProto(pro proto.Message) {
	var err error
	BS.bs.buf, err = proto.Marshal(pro)
	if err != nil {
		BS.bs.buf = BS.bs.buf[:0]
		return
	}
	*BS.bytes = BS.bs.buf
}
func (BS *byteSkin) SetByteString(byteString ByteString) {
	if byteString.buf == nil {
		BS.SetString(byteString.String())
	} else {
		BS.SetBytes(byteString.Bytes())
	}
}
func (BS *byteSkin) View() ByteString {
	return BS.bs
}

// StringSkin ----------------------StringSkin---------------------- StringSkin//
type stringSkin struct {
	str *string
	bs  ByteString
}

func NewStringSkin(str *string) Skin {
	return &stringSkin{
		str: str,
	}
}

func (SS *stringSkin) SetString(str string) {
	bytes := make([]byte, len(str))
	copy(bytes, str)
	SS.bs.str = *(*string)(unsafe.Pointer(&bytes))
	*SS.str = SS.bs.str
}
func (SS *stringSkin) SetBytes(bys []byte) {
	SS.bs.str = string(bys)
	*SS.str = SS.bs.str
}
func (SS *stringSkin) SetProto(pro proto.Message) {
	bytes, err := proto.Marshal(pro)
	if err != nil {
		return
	}
	SS.bs.str = *(*string)(unsafe.Pointer(&bytes))
	*SS.str = SS.bs.str
}
func (SS *stringSkin) SetByteString(byteString ByteString) {
	if byteString.buf == nil {
		SS.SetString(byteString.String())
	} else {
		SS.SetBytes(byteString.buf)
	}
}
func (SS *stringSkin) View() ByteString {
	return SS.bs
}

// protoSkin ----------------------protoSkin---------------------- protoSkin//
type protoSkin struct {
	msg proto.Message
	bs  ByteString
}

func NewProtoSkin(msg proto.Message) Skin {
	if msg == nil {
		return nil
	}
	return &protoSkin{
		msg: msg,
		bs: ByteString{
			buf: []byte{},
		},
	}
}

func (PS *protoSkin) SetString(str string) {
	PS.bs.buf = []byte(str)
	err := proto.Unmarshal(PS.bs.buf, PS.msg)
	if err != nil {
		PS.bs.buf = PS.bs.buf[:0]
		return
	}
}
func (PS *protoSkin) SetBytes(bytes []byte) {
	PS.bs.buf = append(PS.bs.buf[:0], bytes...)
	err := proto.Unmarshal(PS.bs.buf, PS.msg)
	if err != nil {
		PS.bs.buf = PS.bs.buf[:0]
		return
	}
}
func (PS *protoSkin) SetProto(pro proto.Message) {
	var err error
	PS.bs.buf, err = proto.Marshal(pro)
	if err != nil {
		PS.bs.buf = PS.bs.buf[:0]
		return
	}
	_ = proto.Unmarshal(PS.bs.buf, PS.msg)
}
func (PS *protoSkin) SetByteString(byteString ByteString) {
	if byteString.buf == nil {
		PS.SetString(byteString.String())
	} else {
		PS.SetBytes(byteString.Bytes())
	}
}
func (PS *protoSkin) View() ByteString {
	return PS.bs
}

// byteStringSkin ----------------------byteStringSkin---------------------- byteStringSkin//
type byteStringSkin struct {
	bs ByteString
}

func NewByteStringSkin() Skin {
	return &byteStringSkin{
		bs: ByteString{},
	}
}

func (BSS *byteStringSkin) SetString(str string) {
	bytes := make([]byte, len(str))
	copy(bytes, str)
	BSS.bs.str = *(*string)(unsafe.Pointer(&bytes))
}
func (BSS *byteStringSkin) SetBytes(bytes []byte) {
	BSS.bs.str = string(bytes)
}
func (BSS *byteStringSkin) SetProto(pro proto.Message) {
	bytes, err := proto.Marshal(pro)
	if err != nil {
		return
	}
	BSS.bs.str = *(*string)(unsafe.Pointer(&bytes))
}
func (BSS *byteStringSkin) SetByteString(byteString ByteString) {
	if byteString.buf == nil {
		BSS.SetString(byteString.String())
	} else {
		BSS.SetBytes(byteString.Bytes())
	}
}

func (BSS *byteStringSkin) View() ByteString {
	return BSS.bs
}
