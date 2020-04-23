package qiisync

import (
	"fmt"

	"github.com/motemen/go-colorine"
)

var logger = &colorine.Logger{
	Prefixes: colorine.Prefixes{
		"http":  colorine.Verbose,
		"store": colorine.Info,
		"post":  colorine.Info,
		"error": colorine.Error,
		"":      colorine.Verbose,
	},
}

// Logf is a logger that displays logs colorfully.
func Logf(prefix, pattern string, args ...interface{}) {
	logger.Log(prefix, fmt.Sprintf(pattern, args...))
}
