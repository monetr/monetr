package router

type InternalError struct {
	PublicMessage string
	innerError    error
}
