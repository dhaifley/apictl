# apictl
A simple REST API command line interface utility.

## Usage

```
Usage: apictl [<option>] <command> <resource> [<query>]

Options:
  --help = Display this usage message
  --version = Display the command version
  --config.endpoint = Base endpoint URL of the API request
  --config.format = (json|yaml) Format of the command input and output
  --config.headers = Optional, HTTP headers to include with the API request
  --config.tls = Optional, TLS options to use for the API request
  
Commands:
  get
  post, create
  put, update
  patch
  delete
  option, head

Resources:
  Any resource or ID provided by the API. Multiple parameters will be combined
as path segments in the API request.

Query Parameters:
  Any parameters beginning with -- will be sent as query parameters with the API
request. For example, --param=value will be sent as ?param=value.
```

## Examples
```sh
$ apictl --config.format='yaml' \
--config.endpoint='https://example.com/v1/api' \
--config.tls='{"InsecureSkipVerify":true}' \
--config.headers='{"Authorization":["token"]}' \
get users --search='and(email:dev*)' --size=1
```

```sh
created_at: 1721923211
created_by: null
data: null
email: dev@test.com
first_name: null
last_name: null
status: active
updated_at: 1721923211
updated_by: null
user_id: dev@test.com
```

```sh
$ apictl --config.format='yaml' \
--config.endpoint='https://example.com/v1/api' \
--config.tls='{"InsecureSkipVerify":true}' \
--config.headers='{"Authorization":["token"]}' \
get user dev@test.com
```

```sh
created_at: 1721923211
created_by: null
data: null
email: dev@test.com
first_name: null
last_name: null
status: active
updated_at: 1721923211
updated_by: null
user_id: dev@test.com
```

## Building

```sh
$ go build -o apictl main.go
```
