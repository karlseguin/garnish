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

* `Debug()` - Enable debug logging
* `Logger(gc.Logs)` - Use your own logger
* `BytePool(capacity, size uint32)` - A byte pool used whenever your upstreams reply with a Content-Length less than `capacity`. Bytes required will be `capacity * size`.
* `DnsTTL(ttl time.Duration)` - The default TTL to cache dns lookups for. Can be overwritten on a per-upstream basis.

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
* `Slow(d time.Duration)` - the default value to consider a request as being slow. (overwritable on a per-route basis)

#### Authentication

The auth middleware is diabled by default. To enable it, provide the configuration with the `gc.AuthHandler` to use:

```go
config.Auth(authHandler)
...
func authHandler(req *gc.Request) gc.Response {
    ...
}
```

If the configured `AuthHandler` returns a response, that response is immediately returned returned to the client and no further processing is done. In other words, to allow the request to proceed, return `nil` from the handler.

This middleware has no additional configuration.

#### Cache

The cache middleware is disabled by default.

```go
config.Cache().Grace(time.Minute * 2)
```

* `Count(num int)` - The maximum number of responses to keep in the cache
* `Grace(window time.Duration)` - The window to allow a grace response
* `NoSaint()` - Disables saint mode
* `KeyLookup(gc.CacheKeyLookup)` - The function that determines the cache keys to use for this request. A default based on the request's URL + QueryString is used. (overwritable on a per-route basis)
* `PurgeHandler(gc.PurgeHandler)` - The function to call on PURGE requests. No default is provided (it's good to authorize purge requests). If the handler returns a nil response, the request proceeds as normal (thus allowing you to purge the garnish cache and still send the request to the upstream). When a `PurgeHandler` is configured, a route is automatically added to handle any PURGE request.

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

#### Route

At least 1 route must be registered. Routes are registered with a method and associated with an upstream:

```go
config.Route("users").Get("/user/:id").Upstream("test")
```

All routes must have a unique name. Paths support parameters in the form of `:name`, which will be exposed via the `*gc.Request`'s `Param` method. Paths can also end with a wildcard `*`, which acts as a prefix match.

- `Upstream(name string)` - The name of the upstream to send this request to
- `Get(path string)` - The path for a GET method
- `Post(path string)` - The path for a POST method
- `Put(path string)` - The path for a PUT method
- `Delete(path string)` - The path for a DELETE method
- `Purge(path string)` - The path for a PURGE method
- `Patch(path string)` - The path for a PATCH method
- `Head(path string)` - The path for a HEAD method
- `Options(path string)` - The path for a OPTIONS method
- `All(path string)` - The path for a all methods. Can be overwritten for specific methods by specifying the method route first.
- `Slow(t time.Duration)` - Any requests that take longer than `t` to process will be flagged as a slow request by the stats worker. Overwrite's the stat's slow value for this route.
- `CacheTTL(ttl time.Duration)` - The amount of time to cache the response for. Values < 0 will cause the item to never be cached. If the value isn't set, the Cache-Control header received from the upstream will be used.
- `CacheKeyLookup(gc.CacheKeyLookup)` - The function that generates the cache key to use. Overwrites the cache's lookup for this route.


## TODO
- Handle input body draining
- Don't read upstream body until needed (possibly never, thus allowing us to pipe it directly to the ResponseWriter)
- Upstream load balancing
- TCP upstream
- Hydration
- Dispatcher
