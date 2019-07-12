package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type RetryableError interface {
	Err() string
}

type HttpRetryableError struct {
	e error
}

func (re HttpRetryableError) Error() string {
	return re.e.Error()
}

func (re HttpRetryableError) Err() string {
	return re.e.Error()
}

type RequestExecutor interface {
	Do(req *http.Request, timeout time.Duration, entity interface{}) error
}

type defaultExecutor struct{}

func (de *defaultExecutor) Do(req *http.Request, timeout time.Duration, entity interface{}) error {
	client := http.Client{Timeout: timeout}
	if client.Timeout == 0 {
		client.Timeout = time.Second * 10
	}

	r, err := client.Do(req)
	if r != nil {
		defer r.Body.Close()
	}
	if err != nil {
		return err
	}
	if r.StatusCode >= 200 && r.StatusCode < 400 {
		return json.NewDecoder(r.Body).Decode(entity)
	} else if r.StatusCode >= 500 {
		return HttpRetryableError{fmt.Errorf("Error response: %v", r.Status)}
	} else {
		return fmt.Errorf("Error response: %v", r.Status)
	}
}

type Client struct {
	Host       string
	Timeout    time.Duration
	RetryCount int
	Executor   RequestExecutor
}

func (c *Client) Get(path string, headers map[string]string, entity interface{}) error {
	req, err := c.prepareRequestNoBody("GET", path, headers)
	if err != nil {
		return fmt.Errorf("httpclient GET: %v", err)
	}
	return c.doRequestWithRetry(req, entity)
}

func (c *Client) Delete(path string, headers map[string]string, entity interface{}) error {
	req, err := c.prepareRequestNoBody("DELETE", path, headers)
	if err != nil {
		return fmt.Errorf("httpclient DELETE: %v", err)
	}
	return c.doRequestWithRetry(req, entity)
}

func (c *Client) Put(path string, headers map[string]string, inEntity interface{}, outEntity interface{}) error {
	req, err := c.prepareRequestWithBody("PUT", path, headers, inEntity)
	if err != nil {
		return fmt.Errorf("httpclient PUT: %v", err)
	}
	return c.doRequestWithRetry(req, outEntity)
}

func (c *Client) Post(path string, headers map[string]string, inEntity interface{}, outEntity interface{}) error {
	req, err := c.prepareRequestWithBody("POST", path, headers, inEntity)
	if err != nil {
		return fmt.Errorf("httpclient POST: %v", err)
	}
	return c.doRequestWithRetry(req, outEntity)
}

func (c *Client) prepareRequestNoBody(method string, path string, headers map[string]string) (*http.Request, error) {
	return c.makeRequest(method, path, headers, nil)
}

func (c *Client) prepareRequestWithBody(method string, path string, headers map[string]string, entity interface{}) (*http.Request, error) {
	encoded, err := json.Marshal(entity)
	if err != nil {
		return nil, fmt.Errorf("httpclient get: %v", err)
	}

	return c.makeRequest(method, path, headers, bytes.NewBuffer(encoded))
}

func (c *Client) makeURL(path string) (string, error) {
	host := c.Host
	if !strings.HasPrefix(c.Host, "http") {
		host = strings.Join([]string{"http://", c.Host}, "")
	}

	u, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("Unable to parse hostname (%v): %v", c.Host, err)
	}

	parsedPath, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("Unable to parse path (%v): %v", path, err)
	}

	u.Path = parsedPath.Path
	u.RawQuery = parsedPath.RawQuery

	return u.String(), nil
}

func (c *Client) makeRequest(method, path string, headers map[string]string, body io.Reader) (*http.Request, error) {

	u, err := c.makeURL(path)

	req, err := http.NewRequest(method, u, body)
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("could not create request: %v", err)
	}

	return req, nil
}

func (c *Client) doRequestWithRetry(req *http.Request, entity interface{}) error {
	if c.Executor == nil {
		c.Executor = &defaultExecutor{}
	}
	err := c.retryRequest(3*time.Second, func() error {
		return c.Executor.Do(req, c.Timeout, entity)
	})
	return err
}

type httpreqfunc func() error

func (c *Client) retryRequest(sleepTime time.Duration, fn httpreqfunc) error {
	err := fn()
	if err != nil {

		_, isRetryable := err.(RetryableError)
		if !isRetryable {
			return err
		}

		toSleep := sleepTime
		for retries := 0; retries < c.RetryCount; retries++ {
			time.Sleep(toSleep)
			toSleep = toSleep * 2
			err := fn()

			_, isRetryable := err.(RetryableError)
			if err == nil {
				return nil
			} else if !isRetryable {
				return err
			}
		}
	}
	return err
}
