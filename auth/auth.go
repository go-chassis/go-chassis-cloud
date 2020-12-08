package auth

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-chassis/foundation/httpclient"
	"github.com/go-chassis/foundation/security"
	"github.com/go-chassis/go-chassis/v2/security/cipher"
	"github.com/go-chassis/openlog"
)

const (
	HeaderServiceAk      = "X-Service-AK"
	HeaderServiceShaAKSK = "X-Service-ShaAKSK"
	HeaderServiceProject = "X-Service-Project"

	CipherRootEnv   = "CIPHER_ROOT"
	KeytoolAkskFile = "certificate.yaml"

	keyAK      = "servicecomb.credentials.accessKey"
	keySK      = "servicecomb.credentials.secretKey"
	keyProject = "servicecomb.credentials.project"
)

//ErrAuthConfNotExist means the auth config not exist
var ErrAuthConfNotExist = errors.New("auth config is not exist")

func LoadAuth() error {
	err := LoadAkskAuth()
	if err == nil {
		openlog.Info("huawei cloud auth enabled")
		return nil
	}
	if err != ErrAuthConfNotExist {
		openlog.Error(fmt.Sprintf("load ak sk failed: %s", err))
		return err
	}
	openlog.Info("no credential found")
	return nil
}

func getAkskCustomCipher(name string) (security.Cipher, error) {
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
		openlog.Info("CSE uri contains no project")
		return "", nil
	}
	return parts[1], nil
}

// LoadAkskAuth gets the Authentication Mode ak/sk
func LoadAkskAuth() error {
	c, err := getAkskConfig()
	if err != nil {
		return err
	}
	openlog.Info(fmt.Sprintf("huawei cloud auth AK: %s, project: %s", c.AccessKey, c.Project))
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

	httpclient.SignRequest, err = GetShaAKSKSignFunc(c.AccessKey, plainSk, c.Project)
	return err
}
