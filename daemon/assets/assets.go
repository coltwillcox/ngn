package assets

import (
	_ "embed"
)

var (
	//go:embed online.png
	IconOnline []byte
	//go:embed offline.png
	IconOffline []byte
	//go:embed empty.png
	IconEmpty []byte
)
