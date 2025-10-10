package models

//go:generate go run golang.org/x/tools/cmd/stringer@v0.38.0 -type=LinkType -output=link_type.strings.go
type LinkType uint8

const (
	UnknownLinkType LinkType = iota
	PlaidLinkType
	ManualLinkType
	StripeLinkType
)
