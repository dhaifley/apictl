// apictl is a simple command-line utility for making REST API requests.
package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

// Commands
const (
	CmdGet     = http.MethodGet
	CmdCreate  = "CREATE"
	CmdPost    = http.MethodPost
	CmdUpdate  = "UPDATE"
	CmdPut     = http.MethodPut
	CmdPatch   = http.MethodPatch
	CmdDelete  = http.MethodDelete
	CmdOptions = http.MethodOptions
	CmdHead    = http.MethodHead
)

// Formats
const (
	FmtJSON = "json"
	FmtYAML = "yaml"
)

// Args values are used to represent the arguments to the command.
type Args struct {
	Method   string      `json:"method" yaml:"method"`
	Resource string      `json:"resource" yaml:"resource"`
	ID       *string     `json:"id" yaml:"id"`
	Query    *url.Values `json:"query" yaml:"query"`
}

// Config values are used to configure the API requests.
type Config struct {
	Endpoint string       `json:"endpoint" yaml:"endpoint"`
	Headers  *http.Header `json:"headers" yaml:"headers"`
	TLS      *tls.Config  `json:"tls" yaml:"tls"`
	Format   string       `json:"format" yaml:"format"`
}

// Version information.
const Version = "0.0.1"

// Usage details.
const Usage = `Usage: apictl [<option>] <command> <resource> [<id>] [<query>]

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
  --config.tls = Optional, TLS options to use for the API request`

// ParseArgs is used to parse the arguments to the command into the required
// data structures.
func ParseArgs() (*Args, *Config, error) {
	args := &Args{}

	cfg := &Config{}

	cfgMap := map[string]any{}

	for n, arg := range os.Args {
		if n == 0 {
			continue
		}

		if n == 1 {
			switch v := strings.TrimSpace(arg); v {
			case "--version":
				fmt.Println(Version)

				os.Exit(0)
			case "--help", "-?", "-h":
				fmt.Println(Usage)

				os.Exit(0)
			}
		}

		if strings.HasPrefix(arg, "--") {
			switch {
			case strings.HasPrefix(arg, "--config."):
				v := strings.TrimPrefix(arg, "--config.")

				vs := strings.SplitN(v, "=", 2)

				if len(vs) != 2 {
					continue
				}

				vn := strings.ToLower(vs[0])

				var vv any

				if strings.HasPrefix(vs[1], "{") {
					if err := json.Unmarshal([]byte(vs[1]), &vv); err != nil {
						return nil, nil,
							fmt.Errorf("unable to parse config.%s: %w", vn, err)
					}
				} else {
					vv = vs[1]
				}

				cfgMap[vn] = vv
			default:
				p := strings.TrimPrefix(arg, "--")

				ps := strings.SplitN(p, "=", 2)

				if args.Query == nil {
					args.Query = &url.Values{}
				}

				if len(ps) == 2 {
					args.Query.Add(ps[0], ps[1])
				} else {
					args.Query.Set(ps[0], "true")
				}
			}

			continue
		}

		if args.Method == "" {
			switch v := strings.TrimSpace(strings.ToUpper(arg)); v {
			case CmdGet, CmdCreate, CmdPost, CmdUpdate, CmdPut, CmdPatch,
				CmdDelete, CmdOptions, CmdHead:
				args.Method = v
			default:
				return nil, nil, fmt.Errorf("invalid command: %s", v)
			}

			continue
		}

		if args.Resource == "" {
			args.Resource = arg

			continue
		}

		if args.ID == nil {
			args.ID = new(string)
			*args.ID = arg

			continue
		}
	}

	if len(cfgMap) > 0 {
		b, err := json.Marshal(cfgMap)
		if err != nil {
			return nil, nil,
				fmt.Errorf("unable to marshal config map: %w", err)
		}

		if err := json.Unmarshal(b, &cfg); err != nil {
			return nil, nil,
				fmt.Errorf("unable to unmarshal config: %w", err)
		}
	}

	if cfg.Endpoint == "" {
		return nil, nil, fmt.Errorf("missing config.endpoint")
	}

	switch cfg.Format {
	case FmtJSON, FmtYAML:
	case "":
		cfg.Format = FmtJSON
	default:
		return nil, nil, fmt.Errorf("invalid config.format: %s", cfg.Format)
	}

	switch args.Method {
	case CmdCreate:
		args.Method = http.MethodPost
	case CmdUpdate:
		args.Method = http.MethodPut
	}

	return args, cfg, nil
}

// Perform an API request based on provided arguments.
func main() {
	args, cfg, err := ParseArgs()
	if err != nil {
		fmt.Println("ERROR: ", err.Error())

		os.Exit(1)
	}

	ctx := context.Background()

	ur, err := url.Parse(cfg.Endpoint)
	if err != nil {
		fmt.Println("ERROR: invalid endpoint: ", cfg.Endpoint,
			": ", err.Error())

		os.Exit(1)
	}

	ur.Path = path.Join(ur.Path, args.Resource)

	if args.ID != nil {
		ur.Path = path.Join(ur.Path, *args.ID)
	}

	if args.Query != nil {
		ur.RawQuery = args.Query.Encode()
	}

	var buf *bytes.Buffer

	switch args.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Println("ERROR: unable to read input: ", err.Error())

			os.Exit(1)
		}

		if cfg.Format == FmtYAML {
			var r any

			if err := yaml.Unmarshal(b, &r); err != nil {
				fmt.Println("ERROR: unable to parse input YAML: ", err.Error())

				os.Exit(1)
			}

			b, err = json.Marshal(r)
			if err != nil {
				fmt.Println("ERROR: unable to marshal input as JSON: ",
					err.Error())

				os.Exit(1)
			}
		}

		buf = bytes.NewBuffer(b)
	}

	var req *http.Request

	if buf != nil {
		req, err = http.NewRequestWithContext(ctx, args.Method, ur.String(),
			buf)
	} else {
		req, err = http.NewRequestWithContext(ctx, args.Method, ur.String(),
			nil)
	}

	if err != nil || req == nil {
		fmt.Println("ERROR: unable to create request: ", err.Error())

		os.Exit(1)
	}

	if cfg.Headers != nil {
		req.Header = *cfg.Headers
	}

	cli := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: cfg.TLS,
		},
	}

	res, err := cli.Do(req)
	if err != nil {
		fmt.Println("ERROR: unable to perform request: ", err.Error())

		os.Exit(1)
	}

	defer res.Body.Close()

	if args.Method == CmdOptions || args.Method == CmdHead {
		b, err := json.Marshal(res.Header)
		if err != nil {
			fmt.Println("ERROR: unable format response headers: ", err.Error())

			os.Exit(1)
		}

		fmt.Println(res.StatusCode, string(b))

		os.Exit(0)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("ERROR: unable read response body: ", err.Error())

		os.Exit(1)
	}

	ec := 0

	switch {
	case res.StatusCode >= http.StatusInternalServerError:
		ec = 3
	case res.StatusCode >= http.StatusBadRequest:
		ec = 2
	}

	if ec > 0 {
		fmt.Println("ERROR: server error: ", res.StatusCode)
	}

	if len(b) > 0 {
		if cfg.Format == FmtYAML {
			var r any

			if err := json.Unmarshal(b, &r); err != nil {
				fmt.Println("ERROR: unable to parse response JSON: ",
					err.Error())

				os.Exit(1)
			}

			b, err = yaml.Marshal(r)
			if err != nil {
				fmt.Println("ERROR: unable to marshal response as YAML: ",
					err.Error())

				os.Exit(1)
			}
		}

		fmt.Print(string(b))
	}

	os.Exit(ec)
}
