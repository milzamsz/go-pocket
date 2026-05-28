package assets

import "embed"

// FS embeds static frontend assets served by the HTTP layer.
//
//go:embed css/* js/* img/*
var FS embed.FS
