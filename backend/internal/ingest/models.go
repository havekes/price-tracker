package ingest

import "time"

type IngestInput struct {
	CorrespondentName string
	PurchasedAt       time.Time
	Source            string
	ExternalDocID     *string
	RawFileRef        *string
	Items             []IngestItemInput
}

type IngestItemInput struct {
	RawText     string
	DisplayName string
	RawQuantity string
	BaseUnit    string
	TotalPrice  float64
	Currency    string
}
