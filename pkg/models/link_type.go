package models

//go:generate stringer -type=LinkType -output=link_type.strings.go
type LinkType uint8

const (
	UnknownLinkType LinkType = iota
	PlaidLinkType
	ManualLinkType
)
