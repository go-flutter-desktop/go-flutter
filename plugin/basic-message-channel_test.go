package plugin

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	. "github.com/stretchr/testify/assert"
)

// TestBasicMethodChannelStringCodecSend tests the sending a messagen and
// receiving a reply, using a basic message channel with the string codec.
func TestBasicMethodChannelStringCodecSend(t *testing.T) {
	codec := StringCodec{}
	messenger := NewTestingBinaryMessenger()
	messenger.MockSetChannelHandler("ch", func(encodedMessage []byte, r ResponseSender) error {
		message, err := codec.DecodeMessage(encodedMessage)
		if err != nil {
			return errors.Wrap(err, "failed to decode message")
		}
		messageString, ok := message.(string)
		if !ok {
			return errors.New("message is invalid type, expected string")
		}
		reply := messageString + " world"
		encodedReply, err := codec.EncodeMessage(reply)
		if err != nil {
			return errors.Wrap(err, "failed to encode message")
		}
		r.Send(encodedReply)
		return nil
	})
	channel := NewBasicMessageChannel(messenger, "ch", codec)
	reply, err := channel.SendWithReply("hello")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(spew.Sdump(reply))
	replyString, ok := reply.(string)
	if !ok {
		t.Fatal("reply is invalid type, expected string")
	}
	Equal(t, "hello world", replyString)
}

// TestBasicMethodChannelStringCodecHandle tests the handling a messagen and
// sending a reply, using a basic message channel with the string codec.
func TestBasicMethodChannelStringCodecHandle(t *testing.T) {
	codec := StringCodec{}
	messenger := NewTestingBinaryMessenger()
	channel := NewBasicMessageChannel(messenger, "ch", codec)
	channel.HandleFunc(func(message interface{}) (reply interface{}, err error) {
		messageString, ok := message.(string)
		if !ok {
			return nil, errors.New("message is invalid type, expected string")
		}
		reply = messageString + " world"
		return reply, nil
	})
	encodedMessage, err := codec.EncodeMessage("hello")
	if err != nil {
		t.Fatalf("failed to encode message: %v", err)
	}
	encodedReply, err := messenger.MockSend("ch", encodedMessage)
	if err != nil {
		t.Fatal(err)
	}
	reply, err := codec.DecodeMessage(encodedReply)
	if err != nil {
		t.Fatalf("failed to decode reply: %v", err)
	}
	t.Log(spew.Sdump(reply))
	replyString, ok := reply.(string)
	if !ok {
		t.Fatal("reply is invalid type, expected string")
	}
	Equal(t, "hello world", replyString)
}

// TestBasicMethodChannelBinaryCodecSend tests the sending a messagen and
// receiving a reply, using a basic message channel with the binary codec.
func TestBasicMethodChannelBinaryCodecSend(t *testing.T) {
	codec := BinaryCodec{}
	messenger := NewTestingBinaryMessenger()
	messenger.MockSetChannelHandler("ch", func(encodedMessage []byte, r ResponseSender) error {
		message, err := codec.DecodeMessage(encodedMessage)
		if err != nil {
			return errors.Wrap(err, "failed to decode message")
		}
		messageBytes, ok := message.([]byte)
		if !ok {
			return errors.New("message is invalid type, expected []byte")
		}
		reply := append(messageBytes, 0x02)
		encodedReply, err := codec.EncodeMessage(reply)
		if err != nil {
			return errors.Wrap(err, "failed to encode message")
		}
		r.Send(encodedReply)
		return nil
	})
	channel := NewBasicMessageChannel(messenger, "ch", codec)
	reply, err := channel.SendWithReply([]byte{0x01})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(spew.Sdump(reply))
	replyString, ok := reply.([]byte)
	if !ok {
		t.Fatal("reply is invalid type, expected []byte")
	}
	Equal(t, []byte{0x01, 0x02}, replyString)
}

// TestBasicMethodChannelBinaryCodecHandle tests the handling a messagen and
// sending a reply, using a basic message channel with the binary codec.
func TestBasicMethodChannelBinaryCodecHandle(t *testing.T) {
	codec := BinaryCodec{}
	messenger := NewTestingBinaryMessenger()
	channel := NewBasicMessageChannel(messenger, "ch", codec)
	channel.HandleFunc(func(message interface{}) (reply interface{}, err error) {
		messageBytes, ok := message.([]byte)
		if !ok {
			return nil, errors.New("message is invalid type, expected []byte")
		}
		reply = append(messageBytes, 0x02)
		return reply, nil
	})
	encodedMessage, err := codec.EncodeMessage([]byte{0x01})
	if err != nil {
		t.Fatalf("failed to encode message: %v", err)
	}
	encodedReply, err := messenger.MockSend("ch", encodedMessage)
	if err != nil {
		t.Fatal(err)
	}
	reply, err := codec.DecodeMessage(encodedReply)
	if err != nil {
		t.Fatalf("failed to decode reply: %v", err)
	}
	t.Log(spew.Sdump(reply))
	replyString, ok := reply.([]byte)
	if !ok {
		t.Fatal("reply is invalid type, expected []byte")
	}
	Equal(t, []byte{0x01, 0x02}, replyString)
}

func TestBasicMethodChannelNilHandler(t *testing.T) {
	codec := StringCodec{}
	messenger := NewTestingBinaryMessenger()
	channel := NewBasicMessageChannel(messenger, "ch", codec)
	channel.HandleFunc(nil)
	reply, err := messenger.MockSend("ch", []byte("abcd"))
	Nil(t, reply)
	Nil(t, err)
}
func TestBasicMethodChannelNilMockHandler(t *testing.T) {
	codec := StringCodec{}
	messenger := NewTestingBinaryMessenger()
	messenger.MockSetChannelHandler("ch", nil)
	channel := NewBasicMessageChannel(messenger, "ch", codec)
	reply, err := channel.SendWithReply("hello")
	Nil(t, reply)
	NotNil(t, err)
	Equal(t, "failed to send outgoing message: no handler set", err.Error())
}

func TestBasicMethodChannelEncodeFail(t *testing.T) {
	codec := StringCodec{}
	messenger := NewTestingBinaryMessenger()
	channel := NewBasicMessageChannel(messenger, "ch", codec)
	reply, err := channel.SendWithReply(int(42)) // invalid value
	Nil(t, reply)
	NotNil(t, err)
	Equal(t, "failed to encode outgoing message: invalid type provided to message codec: expected message to be of type string", err.Error())
}
