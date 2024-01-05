package qfx

func Validate(data []byte) bool {
	return dataRegex.Match(data)
}
