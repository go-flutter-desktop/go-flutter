package plugin

import (
	"encoding/json"
	"testing"

	. "github.com/stretchr/testify/assert"
)

func TestMethodChannelJSONInvoke(t *testing.T) {
	messenger := NewTestingBinaryMessenger()
	codec := JSONMethodCodec{}
	channel := NewMethodChannel(messenger, "ch", codec)
	messenger.MockSetChannelHandler("ch", func(msg []byte) ([]byte, error) {
		methodCall, err := codec.DecodeMethodCall(msg)
		Nil(t, err)
		NotNil(t, methodCall)
		if methodCall.Method == "sayHello" {
			var greeting string
			err = json.Unmarshal(methodCall.Arguments.(json.RawMessage), &greeting)
			Nil(t, err)
			return codec.EncodeSuccessEnvelope(greeting + " world")
		}
		return codec.EncodeErrorEnvelope("unknown", "", nil)
	})
	result, err := channel.InvokeMethod("sayHello", "hello")
	Nil(t, err)
	Equal(t, json.RawMessage(`"hello world"`), result)

	result, err = channel.InvokeMethod("invalidMethod", "")
	Nil(t, result)
	expectedError := FlutterError{
		Code:    "unknown",
		Message: "",
		Details: json.RawMessage(`null`),
	}
	Equal(t, expectedError, err)
}

//   group('MethodChannel', () {
//     const MessageCodec<dynamic> jsonMessage = JSONMessageCodec();
//     const MethodCodec jsonMethod = JSONMethodCodec();
//     const MethodChannel channel = MethodChannel('ch7', jsonMethod);
//     test('can invoke method and get result', () async {
//       BinaryMessages.setMockMessageHandler(
//         'ch7',
//         (ByteData message) async {
//           final Map<dynamic, dynamic> methodCall = jsonMessage.decodeMessage(message);
//           if (methodCall['method'] == 'sayHello') {
//             return jsonMessage.encodeMessage(<dynamic>['${methodCall['args']} world']);
//           } else {
//             return jsonMessage.encodeMessage(<dynamic>['unknown', null, null]);
//           }
//         },
//       );
//       final String result = await channel.invokeMethod('sayHello', 'hello');
//       expect(result, equals('hello world'));
//     });
//     test('can invoke list method and get result', () async {
//       BinaryMessages.setMockMessageHandler(
//         'ch7',
//         (ByteData message) async {
//           final Map<dynamic, dynamic> methodCall = jsonMessage.decodeMessage(message);
//           if (methodCall['method'] == 'sayHello') {
//             return jsonMessage.encodeMessage(<dynamic>[<String>['${methodCall['args']}', 'world']]);
//           } else {
//             return jsonMessage.encodeMessage(<dynamic>['unknown', null, null]);
//           }
//         },
//       );
//       expect(channel.invokeMethod<List<String>>('sayHello', 'hello'), throwsA(isInstanceOf<TypeError>()));
//       expect(await channel.invokeListMethod<String>('sayHello', 'hello'), <String>['hello', 'world']);
//     });

//     test('can invoke map method and get result', () async {
//       BinaryMessages.setMockMessageHandler(
//         'ch7',
//         (ByteData message) async {
//           final Map<dynamic, dynamic> methodCall = jsonMessage.decodeMessage(message);
//           if (methodCall['method'] == 'sayHello') {
//             return jsonMessage.encodeMessage(<dynamic>[<String, String>{'${methodCall['args']}': 'world'}]);
//           } else {
//             return jsonMessage.encodeMessage(<dynamic>['unknown', null, null]);
//           }
//         },
//       );
//       expect(channel.invokeMethod<Map<String, String>>('sayHello', 'hello'), throwsA(isInstanceOf<TypeError>()));
//       expect(await channel.invokeMapMethod<String, String>('sayHello', 'hello'), <String, String>{'hello': 'world'});
//     });

//     test('can invoke method and get error', () async {
//       BinaryMessages.setMockMessageHandler(
//         'ch7',
//         (ByteData message) async {
//           return jsonMessage.encodeMessage(<dynamic>[
//             'bad',
//             'Something happened',
//             <String, dynamic>{'a': 42, 'b': 3.14},
//           ]);
//         },
//       );
//       try {
//         await channel.invokeMethod<dynamic>('sayHello', 'hello');
//         fail('Exception expected');
//       } on PlatformException catch (e) {
//         expect(e.code, equals('bad'));
//         expect(e.message, equals('Something happened'));
//         expect(e.details, equals(<String, dynamic>{'a': 42, 'b': 3.14}));
//       } catch (e) {
//         fail('PlatformException expected');
//       }
//     });
//     test('can invoke unimplemented method', () async {
//       BinaryMessages.setMockMessageHandler(
//         'ch7',
//         (ByteData message) async => null,
//       );
//       try {
//         await channel.invokeMethod<void>('sayHello', 'hello');
//         fail('Exception expected');
//       } on MissingPluginException catch (e) {
//         expect(e.message, contains('sayHello'));
//         expect(e.message, contains('ch7'));
//       } catch (e) {
//         fail('MissingPluginException expected');
//       }
//     });
//     test('can handle method call with no registered plugin', () async {
//       channel.setMethodCallHandler(null);
//       final ByteData call = jsonMethod.encodeMethodCall(const MethodCall('sayHello', 'hello'));
//       ByteData envelope;
//       await BinaryMessages.handlePlatformMessage('ch7', call, (ByteData result) {
//         envelope = result;
//       });
//       expect(envelope, isNull);
//     });
//     test('can handle method call of unimplemented method', () async {
//       channel.setMethodCallHandler((MethodCall call) async {
//         throw MissingPluginException();
//       });
//       final ByteData call = jsonMethod.encodeMethodCall(const MethodCall('sayHello', 'hello'));
//       ByteData envelope;
//       await BinaryMessages.handlePlatformMessage('ch7', call, (ByteData result) {
//         envelope = result;
//       });
//       expect(envelope, isNull);
//     });
//     test('can handle method call with successful result', () async {
//       channel.setMethodCallHandler((MethodCall call) async => '${call.arguments}, world');
//       final ByteData call = jsonMethod.encodeMethodCall(const MethodCall('sayHello', 'hello'));
//       ByteData envelope;
//       await BinaryMessages.handlePlatformMessage('ch7', call, (ByteData result) {
//         envelope = result;
//       });
//       expect(jsonMethod.decodeEnvelope(envelope), equals('hello, world'));
//     });
//     test('can handle method call with expressive error result', () async {
//       channel.setMethodCallHandler((MethodCall call) async {
//         throw PlatformException(code: 'bad', message: 'sayHello failed', details: null);
//       });
//       final ByteData call = jsonMethod.encodeMethodCall(const MethodCall('sayHello', 'hello'));
//       ByteData envelope;
//       await BinaryMessages.handlePlatformMessage('ch7', call, (ByteData result) {
//         envelope = result;
//       });
//       try {
//         jsonMethod.decodeEnvelope(envelope);
//         fail('Exception expected');
//       } on PlatformException catch (e) {
//         expect(e.code, equals('bad'));
//         expect(e.message, equals('sayHello failed'));
//       } catch (e) {
//         fail('PlatformException expected');
//       }
//     });
//     test('can handle method call with other error result', () async {
//       channel.setMethodCallHandler((MethodCall call) async {
//         throw ArgumentError('bad');
//       });
//       final ByteData call = jsonMethod.encodeMethodCall(const MethodCall('sayHello', 'hello'));
//       ByteData envelope;
//       await BinaryMessages.handlePlatformMessage('ch7', call, (ByteData result) {
//         envelope = result;
//       });
//       try {
//         jsonMethod.decodeEnvelope(envelope);
//         fail('Exception expected');
//       } on PlatformException catch (e) {
//         expect(e.code, equals('error'));
//         expect(e.message, equals('Invalid argument(s): bad'));
//       } catch (e) {
//         fail('PlatformException expected');
//       }
//     });
//   });
