package models

type LinkType string

const (
	UnknownLinkType   LinkType = "unknown"
	PlaidLinkType     LinkType = "plaid"
	ManualLinkType    LinkType = "manual"
	StripeLinkType    LinkType = "stripe"
	LunchFlowLinkType LinkType = "lunch_flow"
)
