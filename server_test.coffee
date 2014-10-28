http = require('http')

handlers =
  plain: (req, res) -> res.end('hello world')
  headers: (req, res) -> res.end(JSON.stringify(req.headers))
  notFound: (req, res) ->
    res.statusCode = 404
    res.end()

server = http.createServer (req, res) ->
  handler = handlers[req.url.substr(1)]
  handler = handlers.notFound unless handler?
  handler(req, res)
server.listen(4005, '127.0.0.1')
