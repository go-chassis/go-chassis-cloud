/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

//Package cse implement client APIs for huawei cloud cse service
package cse

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/go-chassis/foundation/httpclient"
	"github.com/go-chassis/go-chassis/pkg/util/httputil"
)

type Client struct {
	c      *httpclient.Requests
	opts   Options
	Region string
}

func New(opts Options) (*Client, error) {
	ho := &httpclient.Options{
		SSLEnabled: true,
		TLSConfig:  &tls.Config{InsecureSkipVerify: true},
	}
	if opts.Signer != nil {
		ho.SignRequest = opts.Signer
	}
	c, err := httpclient.New(ho)

	return &Client{
		c:    c,
		opts: opts,
	}, err
}

//GetEngineMD return engine information
func (c *Client) GetEngineMD(engineName string) (*EngineMD, error) {
	resp, err := c.c.Get(context.Background(), c.opts.Endpoint+"/cseengine/v1/engine-metadata?name="+engineName, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("bad status:" + resp.Status)
	}
	b := httputil.ReadBody(resp)
	engine := &EngineMD{}
	err = json.Unmarshal(b, engine)
	if err != nil {
		return nil, err
	}
	return engine, nil

}
