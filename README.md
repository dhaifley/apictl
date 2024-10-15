# apictl
A simple REST API command line interface utility.

## Usage

```
Usage: apictl [<option>] <command> <resource> [<id>] [<query>]

Commands:
  get
  post, create
  put, update
  patch
  delete
  option, head

Options:
  --help = Display this usage message
  --config.endpoint = Base endpoint URL of the API request
  --config.format = (json|yaml) Format of the command input and output
  --config.headers = Optional, HTTP headers to include with the API request
  --config.tls = Optional, TLS options to use for the API request
```

## Example
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

## Building

```sh
$ go build -o apictl main.go
```
