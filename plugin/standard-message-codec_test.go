package plugin

import (
	"math"
	"math/big"
	"testing"

	"github.com/davecgh/go-spew/spew"
	. "github.com/stretchr/testify/assert"
)

func TestStandardMessageEncodeIntegers(t *testing.T) {
	// all these values were taken from flutter tests, lowInt and highInt64 must
	// be vars oir the compiler complains about overflowing constant values.
	var lowInt64 = int64(-0x7fffffffffffffff)
	var highInt64 = int64(0x7fffffffffffffff)
	scenarios := []struct {
		value interface{}
		data  []byte
	}{
		{value: int32(-0x7fffffff - 1), data: []byte{3, 0x00, 0x00, 0x00, 0x80}},
		{value: int64(-0x7fffffff - 2), data: []byte{4, 0xff, 0xff, 0xff, 0x7f, 0xff, 0xff, 0xff, 0xff}},
		{value: int32(0x7fffffff), data: []byte{3, 0xff, 0xff, 0xff, 0x7f}},
		{value: int64(0x7fffffff + 1), data: []byte{4, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00}},
		{value: int64(-0x7fffffffffffffff - 1), data: []byte{4, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80}},
		{value: lowInt64 - 2, data: []byte{4, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{value: int64(0x7fffffffffffffff), data: []byte{4, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}},
		{value: highInt64 + 1, data: []byte{4, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80}},
	}
	codec := StandardMessageCodec{}
	for _, s := range scenarios {
		result, err := codec.EncodeMessage(s.value)
		if err != nil {
			t.Fatal(err)
		}
		Equal(t, s.data, result)
	}
}

func TestStandardMessageEncodeSizes(t *testing.T) {
	scenarios := []struct {
		value interface{}
		data  []byte
	}{
		{value: make([]byte, 253), data: append([]byte{8, 253}, make([]byte, 253)...)},
		{value: make([]byte, 254), data: append([]byte{8, 254, 254, 0}, make([]byte, 254)...)},
		{value: make([]byte, 0xffff), data: append([]byte{8, 254, 0xff, 0xff}, make([]byte, 0xffff)...)},
		{value: make([]byte, 0xffff+1), data: append([]byte{8, 255, 0, 0, 1, 0}, make([]byte, 0xffff+1)...)},
	}

	codec := StandardMessageCodec{}

	for _, s := range scenarios {
		result, err := codec.EncodeMessage(s.value)
		if err != nil {
			t.Fatal(err)
		}
		Equal(t, s.data, result)
	}
}

func TestStandardMessageEncodeSimple(t *testing.T) {
	values := []interface{}{
		nil,
		true,
		false,
		int32(7),
		int32(-7),
		int64(98742923489),
		int64(-98742923489),
		int64(9223372036854775807),
		int64(-9223372036854775807),
		big.NewInt(9223372036854775807),
		big.NewInt(-9223372036854775807),
		3.14,
		math.Inf(+1),
		"",
		"hello",
		"special chars >☺😂<",
	}

	codec := StandardMessageCodec{}

	for _, v := range values {
		t.Logf("encoding: %T %v\n", v, v)

		data, err := codec.EncodeMessage(v)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf(spew.Sdump(data))

		v2, err := codec.DecodeMessage(data)
		if err != nil {
			t.Fatal(err)
		}

		Equal(t, v, v2)
	}
}

func TestStandardMessageEncodeNaN(t *testing.T) {
	// Nan != NaN, which causes Equal(..) to give a false negative.

	v := math.NaN()
	codec := StandardMessageCodec{}

	data, err := codec.EncodeMessage(v)
	if err != nil {
		t.Fatal(err)
	}

	v2, err := codec.DecodeMessage(data)
	if err != nil {
		t.Fatal(err)
	}

	v2Float64, ok := v2.(float64)
	if !ok {
		t.Fatal("NaN was not decoded to float64")
	}
	if !math.IsNaN(v2Float64) {
		t.Fatal("NaN was not decoded to NaN")
	}
}

func TestStandardMessageEncodeComposite(t *testing.T) {
	values := []interface{}{
		nil,
		true,
		false,
		int32(-707),
		int64(-7000000007),
		int64(-7000000000000000007),
		big.NewInt(-7000000000000000007),
		float64(-3.14),
		"",
		"hello",
		[]byte{0xBA, 0x5E, 0xBA, 0x11},
		[]int32{-0x7fffffff - 1, 0, 0x7fffffff},
		nil, // ensures the offset of the following list is unaligned.
		[]int64{-0x7fffffffffffffff - 1, 0, 0x7fffffffffffffff},
		nil, // ensures the offset of the following list is unaligned.
		[]float64{
			math.Inf(1),
			math.Inf(-1),
			math.MaxFloat32,
			-math.MaxFloat64,
			0.0,
			-0.0,
		},
		[]interface{}{"nested", []interface{}{}},
		[]interface{}{
			int32(123),
			3.14,
			int64(456),
			"hi",
			[]byte{1, 0xff, 0, 0x0f},
		},
		map[interface{}]interface{}{
			"a": "nested",
			int32(2): []interface{}{
				int32(123),
				453243.4324234,
				int64(456),
				"hi",
				[]byte{1, 0xff, 0, 0x0f},
			},
			nil: map[interface{}]interface{}{
				"foo":     "bar",
				int32(42): int32(43),
				int64(1):  big.NewInt(12345),
			},
		},
		"world",
	}

	// add the whole of values, as last item to values.
	values = append(values, values)

	codec := StandardMessageCodec{}

	for _, v := range values {
		t.Logf("encoding: %T %v\n", v, v)

		data, err := codec.EncodeMessage(v)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf(spew.Sdump(data))

		v2, err := codec.DecodeMessage(data)
		if err != nil {
			t.Fatal(err)
		}

		Equal(t, v, v2)
	}
}

func TestStandardMessageEncodeAlignment(t *testing.T) {
	scenarios := []struct {
		value interface{}
		data  []byte
	}{
		{value: 1.0, data: []byte{6, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xf0, 0x3f}},
	}

	codec := StandardMessageCodec{}

	for _, s := range scenarios {
		result, err := codec.EncodeMessage(s.value)
		if err != nil {
			t.Fatal(err)
		}
		Equal(t, s.data, result)
	}
}
