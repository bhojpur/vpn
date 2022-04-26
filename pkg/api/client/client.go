package client

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/bhojpur/vpn/pkg/api"
	"github.com/bhojpur/vpn/pkg/blockchain"
	"github.com/bhojpur/vpn/pkg/types"
)

type (
	Client struct {
		host       string
		httpClient *http.Client
	}
)

func WithHost(host string) func(c *Client) error {
	return func(c *Client) error {
		c.host = host
		if strings.HasPrefix(host, "unix://") {
			socket := strings.ReplaceAll(host, "unix://", "")
			c.host = "http://unix"
			c.httpClient = &http.Client{
				Transport: &http.Transport{
					DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
						return net.Dial("unix", socket)
					},
				},
			}
		}
		return nil
	}
}

func WithTimeout(d time.Duration) func(c *Client) error {
	return func(c *Client) error {
		c.httpClient.Timeout = d
		return nil
	}
}

func WithHTTPClient(cl *http.Client) func(c *Client) error {
	return func(c *Client) error {
		c.httpClient = cl
		return nil
	}
}

type Option func(c *Client) error

func NewClient(o ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{},
	}
	for _, oo := range o {
		oo(c)
	}
	return c
}

func (c *Client) do(method, endpoint string, params map[string]string) (*http.Response, error) {
	baseURL := fmt.Sprintf("%s%s", c.host, endpoint)

	req, err := http.NewRequest(method, baseURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	q := req.URL.Query()
	for key, val := range params {
		q.Set(key, val)
	}
	req.URL.RawQuery = q.Encode()
	return c.httpClient.Do(req)
}

// Get methods (Services, Users, Files, Ledger, Blockchain, Machines)
func (c *Client) Services() (resp []types.Service, err error) {
	res, err := c.do(http.MethodGet, api.ServiceURL, nil)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return resp, err
	}
	if err = json.Unmarshal(body, &resp); err != nil {
		return resp, err
	}
	return
}

func (c *Client) Files() (data []types.File, err error) {
	res, err := c.do(http.MethodGet, api.FileURL, nil)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return data, err
	}
	if err = json.Unmarshal(body, &data); err != nil {
		return data, err
	}
	return
}

func (c *Client) Users() (data []types.User, err error) {
	res, err := c.do(http.MethodGet, api.UsersURL, nil)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return data, err
	}
	if err = json.Unmarshal(body, &data); err != nil {
		return data, err
	}
	return
}

func (c *Client) Ledger() (data map[string]map[string]blockchain.Data, err error) {
	res, err := c.do(http.MethodGet, api.LedgerURL, nil)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return data, err
	}
	if err = json.Unmarshal(body, &data); err != nil {
		return data, err
	}
	return
}

func (c *Client) Summary() (data types.Summary, err error) {
	res, err := c.do(http.MethodGet, api.SummaryURL, nil)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return data, err
	}
	if err = json.Unmarshal(body, &data); err != nil {
		return data, err
	}
	return
}

func (c *Client) Blockchain() (data blockchain.Block, err error) {
	res, err := c.do(http.MethodGet, api.BlockchainURL, nil)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return data, err
	}
	if err = json.Unmarshal(body, &data); err != nil {
		return data, err
	}
	return
}

func (c *Client) Machines() (resp []types.Machine, err error) {
	res, err := c.do(http.MethodGet, api.MachineURL, nil)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return resp, err
	}
	if err = json.Unmarshal(body, &resp); err != nil {
		return resp, err
	}
	return
}

func (c *Client) GetBucket(b string) (resp map[string]blockchain.Data, err error) {
	res, err := c.do(http.MethodGet, fmt.Sprintf("%s/%s", api.LedgerURL, b), nil)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return resp, err
	}
	if err = json.Unmarshal(body, &resp); err != nil {
		return resp, err
	}
	return
}

func (c *Client) GetBucketKeys(b string) (resp []string, err error) {
	d, err := c.GetBucket(b)
	if err != nil {
		return resp, err
	}
	for k := range d {
		resp = append(resp, k)
	}
	return
}

func (c *Client) GetBuckets() (resp []string, err error) {
	d, err := c.Ledger()
	if err != nil {
		return resp, err
	}
	for k := range d {
		resp = append(resp, k)
	}
	return
}

func (c *Client) GetBucketKey(b, k string) (resp blockchain.Data, err error) {
	res, err := c.do(http.MethodGet, fmt.Sprintf("%s/%s/%s", api.LedgerURL, b, k), nil)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return resp, err
	}

	var r string
	if err = json.Unmarshal(body, &r); err != nil {
		return resp, err
	}

	if err = json.Unmarshal([]byte(r), &r); err != nil {
		return resp, err
	}

	d, err := base64.URLEncoding.DecodeString(r)
	if err != nil {
		return resp, err
	}
	resp = blockchain.Data(string(d))
	return
}

func (c *Client) Put(b, k string, v interface{}) (err error) {
	s := struct{ State string }{}

	dat, err := json.Marshal(v)
	if err != nil {
		return
	}

	d := base64.URLEncoding.EncodeToString(dat)

	res, err := c.do(http.MethodPut, fmt.Sprintf("%s/%s/%s/%s", api.LedgerURL, b, k, d), nil)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(body, &s); err != nil {
		return err
	}

	if s.State != "Announcing" {
		return fmt.Errorf("unexpected state '%s'", s.State)
	}

	return
}

func (c *Client) Delete(b, k string) (err error) {
	s := struct{ State string }{}
	res, err := c.do(http.MethodDelete, fmt.Sprintf("%s/%s/%s", api.LedgerURL, b, k), nil)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, &s); err != nil {
		return err
	}
	if s.State != "Announcing" {
		return fmt.Errorf("unexpected state '%s'", s.State)
	}

	return
}

func (c *Client) DeleteBucket(b string) (err error) {
	s := struct{ State string }{}
	res, err := c.do(http.MethodDelete, fmt.Sprintf("%s/%s", api.LedgerURL, b), nil)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, &s); err != nil {
		return err
	}
	if s.State != "Announcing" {
		return fmt.Errorf("unexpected state '%s'", s.State)
	}

	return
}
