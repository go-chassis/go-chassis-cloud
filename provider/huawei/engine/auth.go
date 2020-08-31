package engine

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chassis/go-chassis-cloud/provider/huawei/env"

	"github.com/go-chassis/foundation/httpclient"
	security2 "github.com/go-chassis/foundation/security"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/core/common"
	"github.com/go-chassis/go-chassis/core/config"
	"github.com/go-chassis/go-chassis/core/config/model"
	"github.com/go-chassis/go-chassis/security/cipher"
	"github.com/go-mesh/openlogging"
	"github.com/huaweicse/auth"
	"gopkg.in/yaml.v2"
)

const (
	cipherRootEnv   = "CIPHER_ROOT"
	keytoolAkskFile = "certificate.yaml"

	keyAK      = "cse.credentials.accessKey"
	keySK      = "cse.credentials.secretKey"
	keyProject = "cse.credentials.project"
)

func loadAuth() error {
	err := loadAkskAuth()
	if err == nil {
		openlogging.GetLogger().Info("Huawei Cloud auth mode: AKSK, source: chassis config")
		return nil
	}
	if err != auth.ErrAuthConfNotExist {
		openlogging.GetLogger().Errorf("Load AKSK failed: %s", err)
		return err
	}

	err = loadPaaSAuth()
	if err == nil {
		openlogging.GetLogger().Warn("Huawei Cloud auth mode: AKSK, source: default secret")
		return nil
	}
	if err != auth.ErrAuthConfNotExist {
		openlogging.GetLogger().Errorf("Get AKSK auth from default secret failed: %s", err)
		return err
	}
	openlogging.GetLogger().Info("no credential found")
	return nil
}

// loadPaaSAuth gets the Authentication Mode ak/sk, token and forms required Auth Headers
func loadPaaSAuth() error {
	h, err := auth.GetAuthHeaderGenerator(
		auth.NewServiceStageRetriever(),
		auth.NewCCERetriever())
	if err != nil {
		return err
	}
	projectFromEnv := env.ProjectName()
	if projectFromEnv != "" {
		openlogging.GetLogger().Infof("huawei cloud project: %s", projectFromEnv)
	}
	httpclient.SignRequest = func(r *http.Request) error {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, vs := range h.GenAuthHeaders() {
			for _, v := range vs {
				r.Header.Add(k, v)
			}
		}
		if projectFromEnv != "" {
			r.Header.Set(auth.HeaderServiceProject, projectFromEnv)
		}
		return nil
	}
	return nil
}

func getAkskCustomCipher(name string) (security2.Cipher, error) {
	f, err := cipher.GetCipherNewFunc(name)
	if err != nil {
		return nil, err
	}
	cipherPlugin := f()
	if cipherPlugin == nil {
		return nil, fmt.Errorf("cipher plugin [%s] invalid", name)
	}
	return cipherPlugin, nil
}

func getProjectFromURI(rawurl string) (string, error) {
	errGetProjectFailed := errors.New("get project from CSE uri failed")
	// rawurl: https://cse.cn-north-1.myhwclouds.com:443
	if rawurl == "" {
		return "", fmt.Errorf("%v, CSE uri empty", errGetProjectFailed)
	}

	u, err := url.Parse(rawurl)
	if err != nil {
		return "", fmt.Errorf("%v, %v", errGetProjectFailed, err)
	}
	parts := strings.Split(u.Host, ".")
	if len(parts) != 4 {
		openlogging.GetLogger().Info("CSE uri contains no project")
		return "", nil
	}
	return parts[1], nil
}

func getAkskConfig() (*model.CredentialStruct, error) {
	// 1, if env CIPHER_ROOT exists, read ${CIPHER_ROOT}/certificate.yaml
	// 2, if env CIPHER_ROOT not exists, read chassis config
	var akskFile string
	if v, exist := os.LookupEnv(cipherRootEnv); exist {
		p := filepath.Join(v, keytoolAkskFile)
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
		c.AccessKey = archaius.GetString(keyAK, "")
		c.SecretKey = archaius.GetString(keySK, "")
		c.Project = archaius.GetString(keyProject, "")
		c.AkskCustomCipher = archaius.GetString(common.AKSKCustomCipher, "")
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
		return nil, auth.ErrAuthConfNotExist
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

// loadAkskAuth gets the Authentication Mode ak/sk
func loadAkskAuth() error {
	c, err := getAkskConfig()
	if err != nil {
		return err
	}
	openlogging.GetLogger().Infof("huawei cloud auth AK: %s, project: %s", c.AccessKey, c.Project)
	plainSk := c.SecretKey
	cipherPluginName := c.AkskCustomCipher
	if cipherPluginName != "" {
		cipherPlugin, err := getAkskCustomCipher(cipherPluginName)
		if err != nil {
			return err
		}
		res, err := cipherPlugin.Decrypt(c.SecretKey)
		if err != nil {
			return fmt.Errorf("decrypt sk failed %v", err)
		}
		plainSk = res
	}

	httpclient.SignRequest, err = auth.GetShaAKSKSignFunc(c.AccessKey, plainSk, c.Project)
	return err
}
