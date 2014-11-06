http = require('http')

hydrateBody = {
  page:1,
  total: 54,
  results: [
    {"!ref": {id: "9001p", type: "product"}},
    {"!ref": {id: "9002p", type: "product"}},
    {"!ref": {id: "9003p", type: "cats"}},
  ]
}

handlers =
  plain: (req, res) -> res.end('hello world')

  cached: (req, res) -> res.end('will it cache?')

  headers: (req, res) -> res.end(JSON.stringify(req.headers))

  hydrate: (req, res) ->
    res.header["X-Hydrate"] = "!ref"
    res.end(JSON.stringify(hydrateBody))

  body: (req, res) ->
    body = ''
    req.on 'data', (d) -> body += d
    req.on 'end', -> res.end(body)

  notFound: (req, res) ->
    res.statusCode = 404
    res.end()

server = http.createServer (req, res) ->
  handler = handlers[req.url.substr(1)]
  handler = handlers.notFound unless handler?
  handler(req, res)
server.listen(4005, '127.0.0.1')
