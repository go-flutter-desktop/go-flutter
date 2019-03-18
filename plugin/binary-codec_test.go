package plugin

import (
	"testing"

	. "github.com/stretchr/testify/assert"
)

func TestBinaryEncodeDecode(t *testing.T) {
	values := []interface{}{
		nil,
		[]byte{},
		[]byte{0, 0, 0, 0},
		[]byte{1, 2, 3, 4},
	}

	codec := BinaryCodec{}

	for _, v := range values {
		data, err := codec.EncodeMessage(v)
		if err != nil {
			t.Fatal(err)
		}
		v2, err := codec.DecodeMessage(data)
		if err != nil {
			t.Fatal(err)
		}
		Equal(t, v, v2)
	}
}
