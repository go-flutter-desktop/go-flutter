#include <stdlib.h>

// INFO: This file was copied from the official GLFW embedder shell
// The author would like to thank the flutter team for their contribution.

// Masks used for UTF-8 to UTF-32 conversion.
static int kTwoByteMask = 0xC0;
static int kThreeByteMask = 0xE0;
static int kFourByteMask = 0xF0;


// Information about the UTF-8 encoded code point.
typedef struct UTF8CodePointInfo {
  // The bit-mask that determines the length of the code point.
  int first_byte_mask;
  // The number of bytes of the code point.
  size_t length;
} UTF8CodePointInfo;

// Creates a [UTF8CodePointInfo] from a given byte. [first_byte] must be the
// first byte in the code point.
UTF8CodePointInfo GetUTF8CodePointInfo(int first_byte) {
  UTF8CodePointInfo byte_info;

  // The order matters. Otherwise, it is possible that comparing against i.e.
  // kThreeByteMask and kFourByteMask could be both true.
  if ((first_byte & kFourByteMask) == kFourByteMask) {
    byte_info.first_byte_mask = 0x07;
    byte_info.length = 4;
  } else if ((first_byte & kThreeByteMask) == kThreeByteMask) {
    byte_info.first_byte_mask = 0x0F;
    byte_info.length = 3;
  } else if ((first_byte & kTwoByteMask) == kTwoByteMask) {
    byte_info.first_byte_mask = 0x1F;
    byte_info.length = 2;
  } else {
    byte_info.first_byte_mask = 0xFF;
    byte_info.length = 1;
  }
  return byte_info;
}

// Queries GLFW for the printable key name given a [key] and [scan_code] and
// converts it to UTF-32. The Flutter framework accepts only one code point,
// therefore, only the first code point will be used. There is unlikely to be
// more than one, but there is no guarantee that it won't happen.
u_int32_t GetUTF32CodePointFromGLFWKey(const char* utf8) {
  // The first byte determines the length of the whole code point.
  const UTF8CodePointInfo byte_info = GetUTF8CodePointInfo(utf8[0]);
  // Tracks how many bits the current byte should shift to the left.
  int shift = byte_info.length - 1;

  const int complement_mask = 0x3F;
  u_int32_t result = 0;

  size_t current_byte_index = 0;
  while (current_byte_index < byte_info.length) {
    const int current_byte = utf8[current_byte_index];
    const int mask =
        current_byte_index == 0 ? byte_info.first_byte_mask : complement_mask;
    current_byte_index++;
    const int bits_to_shift = 6 * shift--;
    result += (current_byte & mask) << bits_to_shift;
  }
  return result;
}
