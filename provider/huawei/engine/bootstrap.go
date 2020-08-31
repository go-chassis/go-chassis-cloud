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

	"github.com/go-chassis/go-chassis-cloud/pkg/client/cse"
	"github.com/go-chassis/go-chassis-cloud/provider/huawei/env"

	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/bootstrap"
	"github.com/go-chassis/go-chassis/core/config"
	"github.com/go-mesh/openlogging"
)

//Init fetch endpoints from engine manager
func Init() error {
	if err := loadAuth(); err != nil {
		return err
	}
	name := archaius.GetString("servicecomb.engine.name", "")
	if name == "" || name == "default" {
		return nil
	}
	openlogging.Info("cse engine name to register to: " + name)
	if env.EngineManagerAddr() == "" {
		return errors.New("engine manager address must be set, when engine name is set")
	}
	c, err := cse.New(cse.Options{Endpoint: env.EngineManagerAddr()})
	if err != nil {
		return err
	}
	md, err := c.GetEngineMD(name)
	if err != nil {
		return err
	}
	config.GlobalDefinition.ServiceComb.Registry.Address = md.CSE.PrivateEndpoint["serviceCenter"]
	config.GlobalDefinition.ServiceComb.Config.Client.ServerURI = md.CSE.PrivateEndpoint["configCenter"]
	config.GlobalDefinition.ServiceComb.Monitor.Client.ServerURI = md.CSE.PrivateEndpoint["dashboardService"]
	openlogging.Info("discover service from engine manager", openlogging.WithTags(
		openlogging.Tags{
			"discovery": config.GlobalDefinition.ServiceComb.Registry.Address,
			"config":    config.GlobalDefinition.ServiceComb.Config.Client.ServerURI,
			"dashboard": config.GlobalDefinition.ServiceComb.Monitor.Client.ServerURI,
		}))
	return nil
}
func init() {
	bootstrap.InstallPlugin("engine_endpoint_fetcher", bootstrap.Func(Init))
}
