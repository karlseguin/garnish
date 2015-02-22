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

You'll then:

1. Configure global settings
2. Configure / enable middlewares
3. Configure specific routes

Finally, you can start garnish:

```go
runtime, err := config.Build()
if err != nil {
  panic(err)
}
garnish.Start(runtime)
```

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

#### Hydration

The Hydration middleware is disabled by default.

Hydration is a form of SSI (Server Side Include) aimed at reducing duplication and latency issues which often come with using a service oriented architecture. The approach is detailed [here](http://openmymind.net/Practical-SOA-Hydration-Part-1/).

```go
config.Hydrate(func (fragment gc.ReferenceFragment) []byte {
	return redis.Get(fragment.String("id"))
}).Header("X-Hydrate")
```

To enable Hydration, a `gc.HydrateLoader` function must be provided. This function is responsible for taking the hydration meta data provided by the upstream and converting it to the actual object. In the above example, the payload is retrieved from Redis.

The provided `gc.ReferenceFragment` exposes a [typed.Typed](https://github.com/karlseguin/typed) object. However, despite this potential flexibility, the current parser is severely limited. Upstreams should provide very basic meta data. Namely, no nested objects are currently supported and values cannot contain a `{}` (for real, teehee).

* `Header(name string)` - The HTTP header the upstream will set to enable hydration against the response.

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
- `Handler(gc.Handler) gc.Reponse` - Provide a custom handler for this route (see handler section)

##### Handers
Each route can have a custom handler. This allows routes to be handled directly in-process, without going to an upstream. For example:

```go
func LoadUser(req *gc.Request, next gc.Middleware) gc.Response {
  id := req.Params().Get("id")
  if len(id) == 0 {
    return gc.NotFoundResponse()
  }
  //todo handle
}
```

By using the `next` argument, it's possible to use a handler AND an upstream:

```go
func DeleteUser(req *gc.Request, next gc.Middleware) gc.Response {
  if isAuthorized(req) == false {
    return gc.Respond(401, "not authorized")
  }
  res := next(req)
  //can manipulate res
  return res
}
```

Because of when it executes (immediately before the upstream), the handler can return a response which takes full advantage of the other middlewares (stats, caching, hydration, authentication, ...)

#### Custom Middleware
You can inject your own middle handler:

```go
config.Insert(gc.BEFORE_STATS, "auth", authHandler)

func authHandler(req *gc.Request, next gc.Middleware) gc.Response{
  if .... {
    return gc.Respond(401, "not authorized")
  }
  return next(req)
}
```

Middlewares can be inserted:

* `BEFORE_STATS`
* `BEFORE_CACHE`
* `BEFORE_HYDRATE`
* `BEFORE_DISPATCH`

## TODO
- Upstream load balancing
- TCP upstream
