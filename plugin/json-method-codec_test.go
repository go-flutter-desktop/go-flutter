package plugin

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONMethodCodecEncodeDecodeSuccessEnvelope(t *testing.T) {
	scenarios := []struct {
		value          interface{}
		decodedMessage json.RawMessage
	}{
		{
			value:          int(42),
			decodedMessage: json.RawMessage("42"),
		},
		{
			value:          float64(3.1415),
			decodedMessage: json.RawMessage("3.1415"),
		},
		{
			value:          "string",
			decodedMessage: json.RawMessage(`"string"`),
		},
		{
			value:          []byte("bytes"),
			decodedMessage: json.RawMessage([]byte(`"Ynl0ZXM="`)),
		},
		{
			value:          []interface{}{"list", 0x0f, "thinks"},
			decodedMessage: json.RawMessage(`["list",15,"thinks"]`),
		},
		{
			value: map[string]interface{}{
				"foo":    "bar",
				"number": 42,
			},
			decodedMessage: json.RawMessage(`{"foo":"bar","number":42}`),
		},
	}

	codec := JSONMethodCodec{}

	for _, scenario := range scenarios {
		message, err := codec.EncodeSuccessEnvelope(scenario.value)
		assert.Nil(t, err)
		decodedMessage, err := codec.DecodeEnvelope(message)
		assert.Nil(t, err)
		assert.Equal(t, scenario.decodedMessage, decodedMessage)
	}

	for i, argument := range scenarios {
		methodName := fmt.Sprintf("metohd.%d", i)
		binaryMessage, err := codec.EncodeMethodCall(MethodCall{
			Method:    methodName,
			Arguments: argument.value,
		})
		assert.Nil(t, err)
		methodCall, err := codec.DecodeMethodCall(binaryMessage)
		assert.Nil(t, err)
		assert.Equal(t, methodName, methodCall.Method)
		assert.Equal(t, argument.decodedMessage, methodCall.Arguments)
	}
}

func TestJSONMethodCodecEncodeDecodeErrorEnvelope(t *testing.T) {
	errorCode := "myErrorCode"
	errorMessage := "myErrorMessage"
	errorDetails := map[string]interface{}{
		"foo":    "bar",
		"number": 42,
	}
	expectedFerr := FlutterError{
		Code:    errorCode,
		Message: errorMessage,
		Details: json.RawMessage(`{"foo":"bar","number":42}`),
	}

	codec := JSONMethodCodec{}

	env, err := codec.EncodeErrorEnvelope(errorCode, errorMessage, errorDetails)
	assert.Nil(t, err)
	assert.NotNil(t, env)
	result, err := codec.DecodeEnvelope(env)
	assert.Nil(t, result)
	assert.Equal(t, expectedFerr, err)
}
