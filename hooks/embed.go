// Package hooks provides embedded hook scripts for CCT
package hooks

import "embed"

//go:embed *.sh
var Scripts embed.FS
