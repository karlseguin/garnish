# Garnish

A reverse proxy

## Usage

Using Garnish comes down to configuring it. You'll most likely want to import two projects:

```go
import(
  "github.com/karlseguin/garnish"
  "github.com/karlseguin/garnish/gc"
)
```

You begin this process by creating a configuration object:

```go
config := garnish.Configure()
```

You'll do four things with the configuration:

1. Configure global settings
2. Configure / enable middlewares
3. Configure specific routes
4. Run Garnish

You can see `example/main.go` for a basic example.

### Global Settings

The most common global setting is:

* `Address(address string)` - the address to listen on

Other settings you likely won't have to change:

* `Debug()` - enable debug logging
* `Logger(gc.Logs)` - use your own logger
* `BytePool(capacity, size uint32)` - a byte pool used whenever your upstreams reply with a Content-Length less than `capacity`. Bytes required will be `capacity * size`.

### Middleware

#### Stats

The stats middleware is disabled by default.

Every minute, the stats middleware writes performance metrics to a file. Stats will generate information about:

1. The runtime
    - # of GCs
    - # of Goroutines
2. Your routes
    - # of hits
    - # of hits by status code (2xx, 4xx, 5xx)
    - # of cache hits
    - # of slow requests
    - 75 percentile load time
    - 95 percentile load time
3. Other
    - Infomration on your byte pool (hits/size/...)

The middleware overwrites the file on each write.

```go
config.Stats().FileName("stats.json").Slow(time.Millisecond * 100)
```

* `FileName(name string)` - the path to save the statistics to
* `Slow(d time.Duration)` - the default value to consider a request as being slow

#### Upstreams

Configures your upstreams. Garnish requires at least 1 upstream to be defined:

```go
config.Upstream("users").Address("http://localhost:4005")
config.Upstream("search").Address("http://localhost:4006")
```

The name given to the upstream is used later whe defining routes.

* `Address(address string)` - The address of the upstream. Must begin with `http://`, `https://` or `unix:/`
* `KeepAlive(count int)` - The number of keepalive connections to maintain with the upstream. Set to 0 to disable
* `DnsCache(ttl time.Duration)` - The length of time to cache the upstream's IP. Even setting this to a short value (1s) can have a significant impact
* `Headers(headers ...string)` - The headers to forward to the upstream
* `Tweaker(tweaker gc.RequestTweaker)` - A RequestTweaker exposes the incoming and outgoing request, allowing you to make any custom changes to the outgoing request.


