package net

import (
	"cache-go/net/pool/slicePool"
	"encoding/binary"
	"math"
)

type (
	CodeC interface {
		//Encode when write
		Encode(bytes []byte) ([]byte, error)
		//Decode when read
		Decode(c Conn) ([]byte, error)
		String() string
	}

	LengthFieldBasedFrameCodec struct {
		order         binary.ByteOrder
		encoderConfig *EncoderConfig
		deCoderConfig *DecoderConfig
	}
	NothingCodec struct {
	}
)

// NewDefaultLengthFieldBasedFrameCodec ------------------------------NewDefaultLengthFieldBasedFrameCodec-----------------------------//
func NewDefaultLengthFieldBasedFrameCodec() CodeC {
	return NewLengthFieldBasedFrameCodec(&EncoderConfig{
		IncludeLengthFieldLength: false,
		LengthAdjustment:         0,
		LengthFieldLength:        64,
	})
}
func NewLengthFieldBasedFrameCodec(encoderConfig *EncoderConfig) CodeC {
	decoderConfig := &DecoderConfig{
		LengthFieldLength:   encoderConfig.LengthFieldLength,
		InitialBytesToStrip: encoderConfig.LengthFieldLength / 8,
		LengthAdjustment:    -encoderConfig.LengthAdjustment,
	}
	if encoderConfig.IncludeLengthFieldLength {
		decoderConfig.LengthAdjustment -= encoderConfig.LengthFieldLength
	}
	return &LengthFieldBasedFrameCodec{
		order:         binary.BigEndian,
		encoderConfig: encoderConfig,
		deCoderConfig: decoderConfig,
	}
}

type EncoderConfig struct {
	LengthFieldLength int64
	//LengthFieldOffset int64
	LengthAdjustment         int64
	IncludeLengthFieldLength bool
}

type DecoderConfig struct {
	LengthFieldLength   int64
	LengthAdjustment    int64
	InitialBytesToStrip int64
}

func (LBC *LengthFieldBasedFrameCodec) String() string {
	return "type:LengthFieldBasedFrameCodec"
}
func (LBC *LengthFieldBasedFrameCodec) Encode(bytes []byte) ([]byte, error) {
	lengthField := int64(len(bytes)) + LBC.encoderConfig.LengthAdjustment
	if LBC.encoderConfig.IncludeLengthFieldLength {
		lengthField += LBC.encoderConfig.LengthFieldLength
	}
	var data []byte
	switch LBC.encoderConfig.LengthFieldLength {
	case 8:
		if lengthField > math.MaxInt8 {
			return nil, ErrorCodeCLengthFieldTooShort
		}
		data = append(data, byte(lengthField))
	case 16:
		if lengthField > math.MaxInt16 {
			return nil, ErrorCodeCLengthFieldTooShort
		}
		data = make([]byte, 2)
		LBC.order.PutUint16(data, uint16(lengthField))
	case 32:
		if lengthField > math.MaxInt32 {
			return nil, ErrorCodeCLengthFieldTooShort
		}
		data = make([]byte, 4)
		LBC.order.PutUint32(data, uint32(lengthField))
	case 64:
		if lengthField > math.MaxInt64 {
			return nil, ErrorCodeCLengthFieldTooShort
		}
		data = make([]byte, 8)
		LBC.order.PutUint64(data, uint64(lengthField))
	default:
		return nil, ErrorInvalidLengthFieldLength
	}
	val := slicePool.Get(len(bytes) + len(data))

	copy(val, data)
	copy(val[len(data):], bytes)
	return val, nil
}

func (LBC *LengthFieldBasedFrameCodec) Decode(c Conn) ([]byte, error) {
	bytes := c.PeekAll()
	if int64(len(bytes)<<3) < LBC.deCoderConfig.LengthFieldLength {
		return nil, ErrorBytesLengthTooShort
	}
	var length int64
	switch LBC.deCoderConfig.LengthFieldLength {
	case 8:
		length = int64(bytes[0])
		bytes = bytes[1:]
	case 16:
		length = int64(LBC.order.Uint16(bytes))
		bytes = bytes[2:]
	case 32:
		length = int64(LBC.order.Uint32(bytes))
		bytes = bytes[4:]
	case 64:
		length = int64(LBC.order.Uint64(bytes))
		bytes = bytes[8:]
	default:
		return nil, ErrorInvalidLengthFieldLength
	}

	length += LBC.deCoderConfig.LengthAdjustment
	if int64(len(bytes)) < length {
		return nil, ErrorBytesLengthTooShort
	}
	val := slicePool.Get(int(length))
	copy(val, bytes[:length])
	c.ShiftN(int(length) + int(LBC.deCoderConfig.InitialBytesToStrip))
	return val, nil
}

// NewNothingCodec  ------------------------------NewNothingCodec-----------------------------//
func NewNothingCodec() CodeC {
	return &NothingCodec{}
}

func (codec *NothingCodec) String() string {
	return "type:NothingCodec"
}

func (codec *NothingCodec) Encode(bytes []byte) ([]byte, error) {
	data := slicePool.Get(len(bytes))
	copy(data, bytes)
	return data, nil
}
func (codec *NothingCodec) Decode(c Conn) ([]byte, error) {
	data := c.PeekAll()
	if len(data) == 0 {
		return nil, ErrorBytesLengthTooShort
	}
	bytes := slicePool.Get(len(data))
	copy(bytes, data)
	c.ShiftN(len(data))
	return bytes, nil
}
