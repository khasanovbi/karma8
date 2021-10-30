package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
)

type dumpRoundTripper struct {
	roundTripper http.RoundTripper
}

func (m *dumpRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	out, err := httputil.DumpRequestOut(request, true)
	if err == nil {
		fmt.Println("========Request========")
		fmt.Println(string(out))
	}

	resp, err := m.roundTripper.RoundTrip(request)
	if err == nil {
		out, dumpErr := httputil.DumpResponse(resp, true)
		if dumpErr == nil {
			fmt.Println("========Response========")
			fmt.Println(string(out))
		}
	}

	return resp, err
}

type client struct {
	client *http.Client
	url    string
	debug  bool
}

func NewClient(url string, debug bool) *client {
	c := http.DefaultClient
	if debug {
		c = &http.Client{
			Transport: &dumpRoundTripper{roundTripper: http.DefaultTransport},
		}
	}

	return &client{
		client: c,
		url:    url,
		debug:  debug,
	}
}

func (m *client) Put(ctx context.Context, filename string, data []byte) error {
	baseURL, err := url.Parse(m.url)
	if err != nil {
		return err
	}

	baseURL.Path = path.Join(baseURL.Path, "file", filename)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, baseURL.String(), bytes.NewReader(data))
	if err != nil {
		return err
	}

	resp, err := m.client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (m *client) Get(ctx context.Context, filename string) ([]byte, error) {
	baseURL, err := url.Parse(m.url)
	if err != nil {
		return nil, err
	}

	baseURL.Path = path.Join(baseURL.Path, "file", filename)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := m.client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func test(endpointURL string, filepath string, debug bool) error {
	c := NewClient(endpointURL, debug)

	f, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("can't open file: %w", err)
	}

	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("can't read data: %w", err)
	}

	filename := uuid.New().String()
	fmt.Println("Send file with name", filename)

	err = c.Put(context.Background(), filename, data)
	if err != nil {
		return fmt.Errorf("can't put file: %w", err)
	}

	storageData, err := c.Get(context.Background(), filename)
	if err != nil {
		return fmt.Errorf("can't get file: %w", err)
	}

	if bytes.Compare(data, storageData) != 0 {
		return errors.New("data from put not equal to data from get")
	}

	return nil
}

func main() {
	endpointURL := flag.String("host", "http://localhost:8080", "endpoint url")
	filepath := flag.String("filepath", "cmd/tester/testdata", "file to upload and return")
	needDumpHTTP := flag.Bool("debug", false, "need to dump http request and response or not")
	flag.Parse()

	err := test(*endpointURL, *filepath, *needDumpHTTP)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}
