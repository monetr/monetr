package ofx

func Validate(data []byte) bool {
	return dataRegex.Match(data)
}
