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
