package vision

// ExtractedItem represents a single line item parsed from a receipt.
type ExtractedItem struct {
	RawText              string  `json:"raw_text"`
	TotalPrice           float64 `json:"total_price"`
	RawQuantity          float64 `json:"raw_quantity"`
	SuggestedProductName string  `json:"suggested_product_name"`
	SuggestedBaseUnit    string  `json:"suggested_base_unit"`
}

// ExtractedReceipt represents the full set of items parsed from a receipt image.
type ExtractedReceipt struct {
	Items []ExtractedItem `json:"items"`
}
