package nginx

import "embed"

// Files exposes the nginx template assets bundled into the binary.
//
//go:embed *.conf
var Files embed.FS
