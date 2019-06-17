package plugin

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"

	"github.com/pkg/errors"
)

const (
	standardMessageTypeNull         = 0
	standardMessageTypeTrue         = 1
	standardMessageTypeFalse        = 2
	standardMessageTypeInt32        = 3
	standardMessageTypeInt64        = 4
	standardMessageTypeBigInt       = 5
	standardMessageTypeFloat64      = 6
	standardMessageTypeString       = 7
	standardMessageTypeByteSlice    = 8
	standardMessageTypeInt32Slice   = 9
	standardMessageTypeInt64Slice   = 10
	standardMessageTypeFloat64Slice = 11
	standardMessageTypeList         = 12
	standardMessageTypeMap          = 13
)

// StandardMessageCodec implements a MessageCodec using the Flutter standard
// binary encoding.
//
// This codec tries to stay compatible with the corresponding
// StandardMessageCodec on the Dart side.
// See: https://docs.flutter.io/flutter/services/StandardMessageCodec-class.html
//
// Supported messages are acyclic values of these forms:
//
//     nil
//     bool
//     byte
//     int32, int64
//     float64
//     *big.Int
//     string
//     []byte, []int32, []int64, []float64
//     []interface{} of supported values
//     map[interface{}]interface{} with supported keys and values
//
// On the Dart side, these values are represented as follows:
//
//     null: null
//     bool: bool
//     byte, int8, int16, int32, int64: int
//     float32, float64: double
//     string: String
//     []byte: Uint8List
//     []int32: Int32List
//     []int64: Int64List
//     []float64: Float64List
//     []interface{}: List
//     map[interface{}]interface{}: Map
//
// *big.Int's are represented in Dart as strings with the
// hexadecimal representation of the integer's value.
//
type StandardMessageCodec struct{}

var _ MessageCodec = StandardMessageCodec{} // compile-time type check

// EncodeMessage encodes message to bytes using the Flutter standard message encoding.
// message is expected to be comprised of supported types. See `type StandardMessageCodec`.
func (s StandardMessageCodec) EncodeMessage(message interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := s.writeValue(&buf, message)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode message")
	}
	return buf.Bytes(), nil
}

// DecodeMessage decodes binary data into a standard message
func (s StandardMessageCodec) DecodeMessage(data []byte) (message interface{}, err error) {
	buf := bytes.NewBuffer(data)
	message, err = s.readValue(buf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode data to message")
	}
	return message, nil
}

// writeSize writes an int representing a size to the specified stream.
// Uses an expanding code of 1 to 5 bytes to optimize for small values.
func (s StandardMessageCodec) writeSize(buf *bytes.Buffer, value int) error {
	if value < 0 {
		return errors.New("invalid size: negative")
	}
	var err error
	if value < 254 {
		// write as byte
		err = buf.WriteByte(byte(value))
		if err != nil {
			return err
		}
	} else if value <= 0xffff {
		// write as uint16
		err = buf.WriteByte(254)
		if err != nil {
			return err
		}
		num := make([]byte, 2)
		endian.PutUint16(num, uint16(value))
		_, err = buf.Write(num)
		if err != nil {
			return err
		}
	} else {
		// write as uint32
		err = buf.WriteByte(255)
		if err != nil {
			return err
		}
		num := make([]byte, 4)
		endian.PutUint32(num, uint32(value))
		_, err = buf.Write(num)
		if err != nil {
			return err
		}
	}
	return nil
}

// writeAlignment writes a number of padding bytes to the specified stream to ensure that
// the next value is aligned to a whole multiple of the specified alignment.
// An example usage with alignment = 8 is to ensure doubles are word-aligned
// in the stream.
func (s StandardMessageCodec) writeAlignment(buf *bytes.Buffer, alignment int) error {
	mod := buf.Len() % alignment
	if mod != 0 {
		_, err := buf.Write(make([]byte, alignment-mod))
		if err != nil {
			return err
		}
	}
	return nil
}

// writeInt16 encodes an int16 to the writer.
// 2 bytes. char in Java.
func (s StandardMessageCodec) writeInt16(w io.Writer, value int16) error {
	return binary.Write(w, endian, value)
}

// writeInt32 encodes an int32 to the writer.
// 4 bytes. int in Java.
func (s StandardMessageCodec) writeInt32(w io.Writer, value int32) error {
	return binary.Write(w, endian, value)
}

// writeInt64 encodes an int64 to the writer.
// 8 bytes. long in Java.
func (s StandardMessageCodec) writeInt64(w io.Writer, value int64) error {
	return binary.Write(w, endian, value)
}

// writeFloat64 encodes a float64 to the writer.
// 8 bytes. double in Java.
func (s StandardMessageCodec) writeFloat64(buf *bytes.Buffer, value float64) error {
	err := s.writeAlignment(buf, 8)
	if err != nil {
		return err
	}
	return binary.Write(buf, endian, value)
}

func (s StandardMessageCodec) writeBigInt(buf *bytes.Buffer, value *big.Int) error {
	return s.writeString(buf, value.Text(16))
}

// writeString encodes a string to the writer.
// First the length is written, then the actual bytes.
func (s StandardMessageCodec) writeString(buf *bytes.Buffer, value string) error {
	err := s.writeSize(buf, len(value))
	if err != nil {
		return err
	}
	_, err = buf.WriteString(value)
	if err != nil {
		return err
	}
	return nil
}

// writeByteSlice encodes a slice of bytes to the writer.
// First the length is written, then the actual bytes.
func (s StandardMessageCodec) writeByteSlice(buf *bytes.Buffer, value []byte) error {
	err := s.writeSize(buf, len(value))
	if err != nil {
		return err
	}
	_, err = buf.Write(value)
	if err != nil {
		return err
	}
	return nil
}

// writeInt32Slice encodes a slice of int32's to the writer.
// First the length is written, then optionally some empty alignment bytes, then the actual bytes.
func (s StandardMessageCodec) writeInt32Slice(buf *bytes.Buffer, value []int32) error {
	err := s.writeSize(buf, len(value))
	if err != nil {
		return err
	}
	err = s.writeAlignment(buf, 4)
	if err != nil {
		return err
	}
	return binary.Write(buf, endian, value)
}

// writeInt64Slice encodes a slice of int64's to the writer.
// First the length is written, then optionally some empty alignment bytes, then the actual bytes.
func (s StandardMessageCodec) writeInt64Slice(buf *bytes.Buffer, value []int64) error {
	err := s.writeSize(buf, len(value))
	if err != nil {
		return err
	}
	err = s.writeAlignment(buf, 8)
	if err != nil {
		return err
	}
	return binary.Write(buf, endian, value)
}

// writeFloat64Slice encodes a slice of float64's to the writer.
// First the length is written, then optionally some empty alignment bytes, then the actual bytes.
func (s StandardMessageCodec) writeFloat64Slice(buf *bytes.Buffer, value []float64) error {
	err := s.writeSize(buf, len(value))
	if err != nil {
		return err
	}
	err = s.writeAlignment(buf, 8)
	if err != nil {
		return err
	}
	return binary.Write(buf, endian, value)
}

func (s StandardMessageCodec) writeList(buf *bytes.Buffer, value []interface{}) error {
	err := s.writeSize(buf, len(value))
	if err != nil {
		return err
	}
	for _, v := range value {
		err = s.writeValue(buf, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s StandardMessageCodec) writeMap(buf *bytes.Buffer, value map[interface{}]interface{}) error {
	err := s.writeSize(buf, len(value))
	if err != nil {
		return err
	}
	for k, v := range value {
		err = s.writeValue(buf, k)
		if err != nil {
			return err
		}
		err = s.writeValue(buf, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// writeValue writes a type discriminator byte followed by the encoded value.
func (s StandardMessageCodec) writeValue(buf *bytes.Buffer, value interface{}) error {
	if value == nil {
		return buf.WriteByte(standardMessageTypeNull)
	}

	var err error
	switch typedValue := value.(type) {
	case bool:
		if typedValue {
			return buf.WriteByte(standardMessageTypeTrue)
		}
		return buf.WriteByte(standardMessageTypeFalse)

	case int32:
		err = buf.WriteByte(standardMessageTypeInt32)
		if err != nil {
			return err
		}
		return s.writeInt32(buf, typedValue)

	case int64:
		err = buf.WriteByte(standardMessageTypeInt64)
		if err != nil {
			return err
		}
		return s.writeInt64(buf, typedValue)

	case float64:
		err = buf.WriteByte(standardMessageTypeFloat64)
		if err != nil {
			return err
		}
		return s.writeFloat64(buf, typedValue)

	case *big.Int:
		err = buf.WriteByte(standardMessageTypeBigInt)
		if err != nil {
			return err
		}
		return s.writeBigInt(buf, typedValue)

	case string:
		err = buf.WriteByte(standardMessageTypeString)
		if err != nil {
			return err
		}
		return s.writeString(buf, typedValue)

	case []byte:
		err = buf.WriteByte(standardMessageTypeByteSlice)
		if err != nil {
			return err
		}
		return s.writeByteSlice(buf, typedValue)

	case []int32:
		err = buf.WriteByte(standardMessageTypeInt32Slice)
		if err != nil {
			return err
		}
		return s.writeInt32Slice(buf, typedValue)

	case []int64:
		err = buf.WriteByte(standardMessageTypeInt64Slice)
		if err != nil {
			return err
		}
		return s.writeInt64Slice(buf, typedValue)

	case []float64:
		err = buf.WriteByte(standardMessageTypeFloat64Slice)
		if err != nil {
			return err
		}
		return s.writeFloat64Slice(buf, typedValue)

	case []interface{}:
		err = buf.WriteByte(standardMessageTypeList)
		if err != nil {
			return err
		}
		return s.writeList(buf, typedValue)

	case map[interface{}]interface{}:
		err = buf.WriteByte(standardMessageTypeMap)
		if err != nil {
			return err
		}
		return s.writeMap(buf, typedValue)

	default:
		return MessageTypeError{fmt.Sprintf("type %T is not supported by StandardMessageCodec", value)}
	}
	// no return statement because each case must return
}

// readSize reads the size, which in standardMessageCodec is encoded as set of
// bytes with variable length.
func (s StandardMessageCodec) readSize(buf *bytes.Buffer) (size int, err error) {
	b, err := buf.ReadByte()
	if err != nil {
		return 0, err
	}
	if b < 254 {
		return int(b), nil
	} else if b == 254 {
		v := endian.Uint16(buf.Next(2))
		return int(v), nil
	}
	v := endian.Uint32(buf.Next(4))
	return int(v), nil
}

// readAlignment reads empty alignment bytes from the buffer. Because
// bytes.Buffer doesn't expose a count of read bytes, which is necessary to
// calculate position and required alignment offset, the originalSize of the
// buffer is passed separately.
func (s StandardMessageCodec) readAlignment(buf *bytes.Buffer, originalSize int, alignment int) {
	position := originalSize - buf.Len()
	mod := position % alignment
	if mod != 0 {
		buf.Next(alignment - mod)
	}
}

func (s StandardMessageCodec) readInt16(r io.Reader) (value int16, err error) {
	err = binary.Read(r, endian, &value)
	return value, err
}

func (s StandardMessageCodec) readInt32(r io.Reader) (value int32, err error) {
	err = binary.Read(r, endian, &value)
	return value, err
}

func (s StandardMessageCodec) readInt64(r io.Reader) (value int64, err error) {
	err = binary.Read(r, endian, &value)
	return value, err
}

func (s StandardMessageCodec) readFloat64(buf *bytes.Buffer, originalSize int) (value float64, err error) {
	s.readAlignment(buf, originalSize, 8)
	err = binary.Read(buf, endian, &value)
	return value, err
}

func (s StandardMessageCodec) readBigInt(buf *bytes.Buffer) (*big.Int, error) {
	bs, err := s.readByteSlice(buf)
	if err != nil {
		return nil, errors.Wrap(err, "reading byte slice for bigint")
	}
	bigint, ok := new(big.Int).SetString(string(bs), 16)
	if !ok {
		return nil, errors.New("invalid binary encoding for bigint")
	}
	return bigint, nil
}

func (s StandardMessageCodec) readString(buf *bytes.Buffer) (string, error) {
	bs, err := s.readByteSlice(buf)
	if err != nil {
		return "", errors.Wrap(err, "reading byte slice for string")
	}
	return string(bs), nil

}

func (s StandardMessageCodec) readByteSlice(buf *bytes.Buffer) (value []byte, err error) {
	length, err := s.readSize(buf)
	if err != nil {
		return nil, errors.Wrap(err, "reading size")
	}
	value = buf.Next(length)
	if len(value) != length {
		return nil, errors.New("message corrupted: not enough bytes in buffer")
	}
	return value, nil
}

func (s StandardMessageCodec) readInt32Slice(buf *bytes.Buffer, originalSize int) ([]int32, error) {
	length, err := s.readSize(buf)
	if err != nil {
		return nil, err
	}
	s.readAlignment(buf, originalSize, 4)
	value := make([]int32, length)
	err = binary.Read(buf, endian, value)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s StandardMessageCodec) readInt64Slice(buf *bytes.Buffer, originalSize int) ([]int64, error) {
	length, err := s.readSize(buf)
	if err != nil {
		return nil, err
	}
	s.readAlignment(buf, originalSize, 8)
	value := make([]int64, length)
	err = binary.Read(buf, endian, value)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s StandardMessageCodec) readFloat64Slice(buf *bytes.Buffer, originalSize int) ([]float64, error) {
	length, err := s.readSize(buf)
	if err != nil {
		return nil, err
	}
	s.readAlignment(buf, originalSize, 8)
	value := make([]float64, length)
	err = binary.Read(buf, endian, value)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s StandardMessageCodec) readList(buf *bytes.Buffer, originalSize int) ([]interface{}, error) {
	length, err := s.readSize(buf)
	if err != nil {
		return nil, err
	}
	list := make([]interface{}, 0, length)
	for i := 0; i < length; i++ {
		value, err := s.readValueAligned(buf, originalSize)
		if err != nil {
			return nil, err
		}
		list = append(list, value)
	}
	return list, nil
}

func (s StandardMessageCodec) readMap(buf *bytes.Buffer, originalSize int) (map[interface{}]interface{}, error) {
	length, err := s.readSize(buf)
	if err != nil {
		return nil, err
	}
	m := make(map[interface{}]interface{})
	for i := 0; i < length; i++ {
		key, err := s.readValueAligned(buf, originalSize)
		if err != nil {
			return nil, err
		}
		value, err := s.readValueAligned(buf, originalSize)
		if err != nil {
			return nil, err
		}
		m[key] = value
	}
	return m, nil
}

func (s StandardMessageCodec) readValue(buf *bytes.Buffer) (value interface{}, err error) {
	originalSize := buf.Len()
	return s.readValueAligned(buf, originalSize)
}

func (s StandardMessageCodec) readValueAligned(buf *bytes.Buffer, originalSize int) (value interface{}, err error) {
	valueType, err := buf.ReadByte()
	if err != nil {
		return nil, errors.Wrap(err, "reading value type")
	}

	switch valueType {
	case standardMessageTypeNull:
		return nil, nil

	case standardMessageTypeTrue:
		return true, nil

	case standardMessageTypeFalse:
		return false, nil

	case standardMessageTypeInt32:
		return s.readInt32(buf)

	case standardMessageTypeInt64:
		return s.readInt64(buf)

	case standardMessageTypeFloat64:
		return s.readFloat64(buf, originalSize)

	case standardMessageTypeBigInt:
		return s.readBigInt(buf)

	case standardMessageTypeString:
		return s.readString(buf)

	case standardMessageTypeByteSlice:
		return s.readByteSlice(buf)

	case standardMessageTypeInt32Slice:
		return s.readInt32Slice(buf, originalSize)

	case standardMessageTypeInt64Slice:
		return s.readInt64Slice(buf, originalSize)

	case standardMessageTypeFloat64Slice:
		return s.readFloat64Slice(buf, originalSize)

	case standardMessageTypeList:
		return s.readList(buf, originalSize)

	case standardMessageTypeMap:
		return s.readMap(buf, originalSize)

	default:
		return nil, errors.New("invalid message value type")
	}
}
