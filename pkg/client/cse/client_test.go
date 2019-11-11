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

package cse_test

import (
	"github.com/go-chassis/foundation/httpclient"
	"github.com/go-chassis/go-chassis-cloud/pkg/client/cse"
	"github.com/huaweicse/auth"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	c, err := cse.New(cse.Options{Endpoint: "https://cse.cn-north-1.myhuaweicloud.com"})
	assert.NoError(t, err)
	httpclient.SignRequest, err = auth.GetSignFunc("xxx", "xxx", "cn-north-1")
	assert.NoError(t, err)
	os.Setenv("HTTP_DEBUG", "1")
	engine, err := c.GetEngineMD("default")
	assert.NoError(t, err)
	t.Log(engine.CSE.PrivateEndpoint)
	t.Log(engine.CSE.PublicEndpoint)
}
