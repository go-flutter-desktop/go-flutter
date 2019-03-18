package plugin

import (
	"testing"

	. "github.com/stretchr/testify/assert"
)

func TestStringEncodeDecode(t *testing.T) {
	values := []interface{}{
		nil,
		"",
		"hello",
		"special chars >☺😂<",
	}

	codec := StringCodec{}

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

func TestStringEncodeFail(t *testing.T) {
	codec := StringCodec{}

	// invalid type
	_, err := codec.EncodeMessage(int(42))
	NotNil(t, err)

	// invalid 2-octet utf-8 sequence
	_, err = codec.EncodeMessage("\xc3\x28")
	NotNil(t, err)
}

func TestStringDecodeFail(t *testing.T) {
	codec := StringCodec{}

	// invalid 2-octet utf-8 sequence
	_, err := codec.DecodeMessage([]byte("\xc3\x28"))
	NotNil(t, err)
}
