package garnish

import (
	"github.com/karlseguin/garnish/middleware/stats"
)

var (
	Stats = stats.Configure()
)
