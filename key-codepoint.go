package flutter

// Queries GLFW for the printable key name given a [key] and [scan_code] and
// converts it to UTF-32. The Flutter framework accepts only one code point,
// therefore, only the first code point will be used. There is unlikely to be
// more than one, but there is no guarantee that it won't happen.
func codepointFromGLFWKey(utf8 ...rune) (result int32) {
	return int32(utf8[0])
}
