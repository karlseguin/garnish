package stats

import (
	"runtime"
	"runtime/debug"
)

var (
	gcstats    = new(debug.GCStats)
	runtimeMap = map[string]int64{
		"goroutines": 0,
		"gcs":        0,
		"pausetotal": 0,
		"lastpause":  0,
	}
)

func Runtime() map[string]int64 {
	debug.ReadGCStats(gcstats)
	runtimeMap["goroutines"] = int64(runtime.NumGoroutine())
	runtimeMap["gcs"] = gcstats.NumGC
	runtimeMap["pausetotal"] = int64(gcstats.PauseTotal)
	if len(gcstats.Pause) > 1 {
		runtimeMap["lastpause"] = int64(gcstats.Pause[0])
	} else {
		runtimeMap["lastpause"] = 0
	}
	return runtimeMap
}
