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
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	hws_cloud "github.com/huaweicse/auth/third_party/forked/datastream/aws"
	"net/http"
)

//SignRequest inject auth related header and sign this request so that this request can access to huawei cloud
type SignRequest func(*http.Request) error

// GetSignFunc sets and initializes the ak/sk auth func
func GetSignFunc(ak, sk, project string) (SignRequest, error) {
	s := &hws_cloud.Signer{
		AccessKey: ak,
		SecretKey: sk,
		Service:   "",
		Region:    "",
	}

	shaAKSKSignFunc, err := GetShaAKSKSignFunc(ak, sk, project)
	if err != nil {
		return nil, err
	}

	return func(r *http.Request) error {
		if err := shaAKSKSignFunc(r); err != nil {
			return err
		}
		return s.Sign(r)
	}, nil
}

// GetShaAKSKSignFunc sets and initializes the ak/sk auth func
func GetShaAKSKSignFunc(ak, sk, project string) (SignRequest, error) {
	shaAKSK, err := genShaAKSK(sk, ak)
	if err != nil {
		return nil, err
	}

	return func(r *http.Request) error {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set(HeaderServiceAk, ak)
		r.Header.Set(HeaderServiceShaAKSK, shaAKSK)
		r.Header.Set(HeaderServiceProject, project)
		return nil
	}, nil
}

func genShaAKSK(key string, data string) (string, error) {
	h := hmac.New(sha256.New, []byte(key))
	if _, err := h.Write([]byte(data)); err != nil {
		return "", err
	}
	b := h.Sum(nil)
	shaaksk := ""
	for _, j := range b {
		shaaksk = shaaksk + fmt.Sprintf("%02x", j)
	}
	return shaaksk, nil
}
