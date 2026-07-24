package ingest

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/havekes/price-tracker/internal/db"
	"github.com/havekes/price-tracker/internal/normalize"
	"github.com/havekes/price-tracker/internal/store"
)

func PersistReceipt(ctx context.Context, storeQuerier store.Querier, tx *sql.Tx, input IngestInput) error {
	q := storeQuerier.WithTx(tx)

	// Resolve or create correspondent
	corr, err := q.GetCorrespondentByName(ctx, input.CorrespondentName)
	var corrID int64
	if err != nil {
		if err == sql.ErrNoRows {
			newCorr, err := q.CreateCorrespondent(ctx, input.CorrespondentName)
			if err != nil {
				return fmt.Errorf("failed to create correspondent: %w", err)
			}
			corrID = newCorr.ID
		} else {
			return fmt.Errorf("failed to get correspondent: %w", err)
		}
	} else {
		corrID = corr.ID
	}

	// Insert Receipt
	var extDocID sql.NullString
	if input.ExternalDocID != nil {
		extDocID = sql.NullString{String: *input.ExternalDocID, Valid: true}
	}
	var rawFileRef sql.NullString
	if input.RawFileRef != nil {
		rawFileRef = sql.NullString{String: *input.RawFileRef, Valid: true}
	}

	receiptParams := db.CreateReceiptParams{
		CorrespondentID: corrID,
		PurchasedAt:     input.PurchasedAt,
		Source:          input.Source,
		ExternalDocID:   extDocID,
		RawFileRef:      rawFileRef,
	}

	receipt, err := q.CreateReceipt(ctx, receiptParams)
	if err != nil {
		return fmt.Errorf("failed to create receipt: %w", err)
	}

	for _, item := range input.Items {
		// Resolve or create Product
		prod, err := q.GetProductByName(ctx, item.DisplayName)
		var prodID int64
		if err != nil {
			if err == sql.ErrNoRows {
				stdBaseUnit := normalize.StandardizeUnit(item.BaseUnit)
				newProd, err := q.CreateProduct(ctx, db.CreateProductParams{
					DisplayName: item.DisplayName,
					BaseUnit:    stdBaseUnit,
				})
				if err != nil {
					return fmt.Errorf("failed to create product %q: %w", item.DisplayName, err)
				}
				prodID = newProd.ID
			} else {
				return fmt.Errorf("failed to get product %q: %w", item.DisplayName, err)
			}
		} else {
			prodID = prod.ID
		}

		// Parse Quantity
		var val float64
		var unit string
		var qtyValue sql.NullFloat64
		var qtyUnit sql.NullString

		if item.RawQuantity != "" {
			val, unit, err = normalize.ParseQuantity(item.RawQuantity)
			if err == nil {
				qtyValue = sql.NullFloat64{Float64: val, Valid: true}
				qtyUnit = sql.NullString{String: unit, Valid: true}
			}
		}

		// Insert RawItem
		rawItemParams := db.CreateRawItemParams{
			ReceiptID:     receipt.ID,
			ProductID:     sql.NullInt64{Int64: prodID, Valid: true},
			RawText:       item.RawText,
			RawQuantity:   sql.NullString{String: item.RawQuantity, Valid: item.RawQuantity != ""},
			QuantityValue: qtyValue,
			QuantityUnit:  qtyUnit,
		}

		rawItem, err := q.CreateRawItem(ctx, rawItemParams)
		if err != nil {
			return fmt.Errorf("failed to create raw item %q: %w", item.RawText, err)
		}

		// Calculate Unit Price
		var unitPrice sql.NullFloat64
		if qtyValue.Valid && qtyUnit.Valid {
			// Product's base unit may be different
			actualBaseUnit := item.BaseUnit
			if prod.ID != 0 {
				actualBaseUnit = prod.BaseUnit
			} else {
				actualBaseUnit = normalize.StandardizeUnit(item.BaseUnit)
			}

			up, err := normalize.CalculateUnitPrice(item.TotalPrice, qtyValue.Float64, qtyUnit.String, actualBaseUnit)
			if err == nil {
				unitPrice = sql.NullFloat64{Float64: up, Valid: true}
			}
		}

		// Insert PriceRecord
		priceParams := db.CreatePriceRecordParams{
			RawItemID:  rawItem.ID,
			TotalPrice: item.TotalPrice,
			UnitPrice:  unitPrice,
			Currency:   item.Currency,
		}

		if priceParams.Currency == "" {
			priceParams.Currency = "USD"
		}

		_, err = q.CreatePriceRecord(ctx, priceParams)
		if err != nil {
			return fmt.Errorf("failed to create price record for raw item %d: %w", rawItem.ID, err)
		}
	}

	return nil
}
