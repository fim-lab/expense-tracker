package expensetracker

import "embed"

// StaticAssets embeds the frontend directory.
//go:embed all:frontend
var StaticAssets embed.FS