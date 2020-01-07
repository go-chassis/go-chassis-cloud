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

package engine

import (
	"errors"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/core/config"
	"github.com/go-mesh/openlogging"
	"strings"

	"github.com/go-chassis/go-chassis-cloud/pkg/client/cse"
	"github.com/go-chassis/go-chassis-cloud/provider/huawei/env"
	"github.com/go-chassis/go-chassis/bootstrap"
)

var (
	ErrEmptyRegion  = errors.New("env " + env.ENVRegion + " is empty, please manually set it")
	ErrNoEngineName = errors.New("engine name is empty")
)

//Init fetch endpoints from engine manager
func Init() error {
	if err := loadAuth(); err != nil {
		return err
	}
	region := env.RegionName()
	if region == "" {
		return ErrEmptyRegion
	}
	name := archaius.GetString("servicecomb.engine.name", "")
	if name == "" {
		return ErrNoEngineName
	}
	ep := strings.Join([]string{"https://cse", region, "myhuaweicloud.com"}, ".")
	c, err := cse.New(cse.Options{Endpoint: ep})
	if err != nil {
		return err
	}
	md, err := c.GetEngineMD(name)
	if err != nil {
		return err
	}
	config.GlobalDefinition.Cse.Service.Registry.Address = md.CSE.PrivateEndpoint["serviceCenter"]
	config.GlobalDefinition.Cse.Service.Registry.ServiceDiscovery.Address = config.GlobalDefinition.Cse.Service.Registry.Address
	config.GlobalDefinition.Cse.Config.Client.ServerURI = md.CSE.PrivateEndpoint["configCenter"]
	config.GlobalDefinition.Cse.Monitor.Client.ServerURI = md.CSE.PrivateEndpoint["dashboardService"]
	openlogging.Info("discover service from engine manager", openlogging.WithTags(
		openlogging.Tags{
			"discovery": config.GlobalDefinition.Cse.Service.Registry.Address,
			"config":    config.GlobalDefinition.Cse.Config.Client.ServerURI,
			"dashboard": config.GlobalDefinition.Cse.Monitor.Client.ServerURI,
		}))
	return nil
}
func init() {
	bootstrap.InstallPlugin("huaweiauth", bootstrap.Func(Init))
	bootstrap.InstallPlugin("engine_endpoint_fetcher", bootstrap.Func(Init))
}
