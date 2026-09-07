package models

type LinkType string

const (
	UnknownLinkType   LinkType = "unknown"
	PlaidLinkType     LinkType = "plaid"
	ManualLinkType    LinkType = "manual"
	LunchFlowLinkType LinkType = "lunch_flow"
)
