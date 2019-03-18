package plugin

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// JSONMethodCodec implements a MethodCodec using JSON for message encoding.
type JSONMethodCodec struct{}

var _ MethodCodec = JSONMethodCodec{}

// EncodeMethodCall encodes the MethodCall into binary Returns an error on
// invalid MethodCall arguments.
func (j JSONMethodCodec) EncodeMethodCall(methodCall MethodCall) (data []byte, err error) {
	jmc := struct {
		Method string      `json:"method"`
		Args   interface{} `json:"args"`
	}{
		Method: methodCall.Method,
		Args:   methodCall.Arguments,
	}
	return json.Marshal(&jmc)
}

// DecodeMethodCall decodes the MethodCall from binary. Nore that the MethodCall
// arguments are not fully parsed, they are always a json.RawMessage and must be
// decoded by the MethodHandler. Returns an error on invalid data.
func (j JSONMethodCodec) DecodeMethodCall(data []byte) (methodCall MethodCall, err error) {
	jmc := struct {
		Method string          `json:"method"`
		Args   json.RawMessage `json:"args"`
	}{}
	err = json.Unmarshal(data, &jmc)
	if err != nil {
		return MethodCall{}, errors.Wrap(err, "failed to decode json method call")
	}
	mc := MethodCall{
		Method:    jmc.Method,
		Arguments: jmc.Args,
	}
	return mc, nil
}

// EncodeSuccessEnvelope encodes a successful result into a binary envelope. The
// result value must be encodable in JSON.
func (j JSONMethodCodec) EncodeSuccessEnvelope(result interface{}) (data []byte, err error) {
	return json.Marshal([]interface{}{result})
}

// EncodeErrorEnvelope encodes an error result into a binary envelope. The
// specified error code, human-readable error message, and error details
// correspond to the fields of Flutter's PlatformException.
func (j JSONMethodCodec) EncodeErrorEnvelope(code string, message string, details interface{}) (data []byte, err error) {
	return json.Marshal([]interface{}{code, message, details})
}

// DecodeEnvelope decodes the specified result envelope from binary. Returns a
// FlutterError as error if provided envelope represents an error, otherwise
// returns the result as a json.RawMessage
func (j JSONMethodCodec) DecodeEnvelope(envelope []byte) (result interface{}, err error) {
	fields := []json.RawMessage{}
	err = json.Unmarshal(envelope, &fields)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode envelope")
	}

	if len(fields) == 1 {
		return fields[0], nil
	}
	if len(fields) == 3 {
		ferr := FlutterError{
			Details: fields[2],
		}
		err = json.Unmarshal(fields[0], &ferr.Code)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode field 'code' from json error envelope")
		}
		err = json.Unmarshal(fields[1], &ferr.Message)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode field 'message' from json error envelope")
		}
		return nil, ferr
	}
	return nil, errors.New("invalid JSON envelope")
}
