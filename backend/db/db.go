package db

import (
	_ "embed"
)

// Schema is the DDL SQL that creates all tables.
// Embedded from schema.sql for use in migration and tests.
//
//go:embed schema.sql
var Schema string
