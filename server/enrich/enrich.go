package enrich

import (
	"context"
	"time"
)

type Direction string

const (
	OutflowDirection Direction = "outflow"
	InflowDirection  Direction = "inflow"
)

type Confidence uint8

const (
	UnknownConfidence  Confidence = 0
	LowConfidence      Confidence = 1
	MediumConfidence   Confidence = 2
	HighConfidence     Confidence = 3
	VeryHighConfidence Confidence = 4
)

type Item struct {
	// ItemId is the unique identifier for the specific transaction that you want
	// to enrich. It will be used to identify an item in the enrichment result.
	ItemId string
	// AccountHolderId is used to separate transactions by the account they belong
	// to. This should not be personally identifiable but should instead by an
	// internal Id so that data that belongs to the same account from the same
	// data source are grouped together in the enrichment process.
	AccountHolderId string
	// Amount should be the absolute amount of the transaction in the smallest
	// denomenator of the currency the transaction is in. For example; in USD this
	// should always be positive cents.
	Amount int64
	// CurrencyCode should be the ISO currency code for the transaction being
	// processed.
	CurrencyCode string
	// Direction should specify whether the transaction was a withdrawal (outflow)
	// or if the transaction was a deposit (inflow).
	Direction Direction
	// Description should be the original description as provided by the data
	// source. Original transaction name is preferred if possible.
	Description string
	// Date should be the date of the transaction. The time will be truncated when
	// processed by the enrichment process.
	Date time.Time
}

type Enrichment struct {
	// ItemId is the unique identifier for the specific transaction that you
	// initially provided as input to the processor.
	ItemId string
	// Description will be a sanitized version of the original description based
	// on the enrichment processor's data. If no enrichment was performed at all
	// or the original description was already sanitized then it is possible for
	// this to just be the original transaction description.
	Description string
	// Merchant represents the counterparty entity of the transaction. This will
	// not always be present depending on how well the processor is able to enrich
	// a transaction.
	Merchant *Entity
	// Intermediares are entities that were part of the transaction processing. An
	// example of an intermediary would be Privacy or PayPal.
	Intermediaries []Entity
	// If a category can be determined from the data then it will be included
	// here.
	Category *Category
}

type Entity struct {
	// MerchantId might be provided if the enrichment processor supports uniquely
	// identifying merchants.
	EntityId *string
	// Name is the sanitized name of the merchant as inferred by the provided
	// transaction data.
	Name string
	// Website will be a URL for the merchant if one is available.
	Website *string
	// Logo will be a URL compatible path for the merchant if one is available. It
	// is possible for this to be a Base64 encoded image as a data url though.
	Logo *string
	// Confidence will be returned with an entity to indicate how likely the data
	// included for this entity is accurate. If confidence is not provided by the
	// enrichment processor then this will always be UnknownConfidence.
	Confidence Confidence
}

type Category struct {
	// Primary will be the primary high level category from Plaid's taxonomy.
	Primary string
	// Detailed will be the specific category within the primary from Plaid's
	// taxonomy.
	Detailed string
	// Confidence will be returned with the category to indicate how likely this
	// category is correct. If the enrichment processor does not provide a
	// confidence indicator then this will always be UnknownConfidence.
	Confidence Confidence
}

type EnrichmentProcessor interface {
	// Enrich will take an array of transaction items (maximum of 100) and enrich
	// them synchronously. It will return an array of enriched transactions. There
	// will always be at least one enrichment per transaction but some details on
	// the enrichment may be missing if enrichment was not possible for that
	// transaction.
	Enrich(ctx context.Context, input []Item) ([]Enrichment, error)
}
