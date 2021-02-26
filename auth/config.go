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

package auth

import (
	"errors"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis-cloud/provider/huawei/env"
	"github.com/go-chassis/go-chassis/v2/core/common"
	"github.com/go-chassis/go-chassis/v2/core/config"
	"github.com/go-chassis/go-chassis/v2/core/config/model"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

func getAkskConfig() (*model.CredentialStruct, error) {
	// 1, if env CIPHER_ROOT exists, read ${CIPHER_ROOT}/certificate.yaml
	// 2, if env CIPHER_ROOT not exists, read chassis config
	var akskFile string
	if v, exist := os.LookupEnv(CipherRootEnv); exist {
		p := filepath.Join(v, KeytoolAkskFile)
		if _, err := os.Stat(p); err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
		} else {
			akskFile = p
		}
	}

	c := &model.CredentialStruct{}
	if akskFile == "" {
		c.AccessKey = archaius.GetString(keyAKV2, archaius.GetString(keyAK, ""))
		c.SecretKey = archaius.GetString(keySKV2, archaius.GetString(keySK, ""))
		c.Project = archaius.GetString(keyProjectV2, archaius.GetString(keyProject, ""))
		c.AkskCustomCipher = archaius.GetString(common.AKSKCustomCipher, archaius.GetString("cse.credentials.akskCustomCipher", ""))
	} else {
		yamlContent, err := ioutil.ReadFile(akskFile)
		if err != nil {
			return nil, err
		}
		globalConf := &model.GlobalCfg{}
		err = yaml.Unmarshal(yamlContent, globalConf)
		if err != nil {
			return nil, err
		}
		c = &(globalConf.ServiceComb.Credentials)
	}
	if c.AccessKey == "" && c.SecretKey == "" {
		return nil, ErrAuthConfNotExist
	}
	if c.AccessKey == "" || c.SecretKey == "" {
		return nil, errors.New("ak or sk is empty")
	}

	// 1, use project of env PAAS_PROJECT_NAME
	// 2, use project in the credential config
	// 3, use project in cse uri contain
	// 4, use project "default"
	if v := env.ProjectName(); v != "" {
		c.Project = v
	}
	if c.Project == "" {
		project, err := getProjectFromURI(config.GetRegistratorAddress())
		if err != nil {
			return nil, err
		}
		if project != "" {
			c.Project = project
		} else {
			c.Project = common.DefaultValue
		}
	}
	return c, nil
}
