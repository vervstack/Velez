package security

import (
	"bytes"
	"encoding/json"
	"os"
	"path"
	"sync"

	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"
	rtb "go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/config"
)

const defaultPath = "/tmp/velez/private-keys.json"

type Manager struct {
	buildPath string
	keys      PrivateKeys

	sync.Once
}

type PrivateKeys struct {
	Velez     string
	Matreshka string
	Headscale string
}

func NewSecurityManager(cfg config.Config) *Manager {
	if cfg.Environment.CustomPassToKey == "" {
		cfg.Environment.CustomPassToKey = defaultPath
	}

	return &Manager{
		buildPath: cfg.Environment.CustomPassToKey,
		keys:      PrivateKeys{},
	}
}

func (s *Manager) Start() error {
	var err error
	s.Once.Do(func() {
		err = s.start()
	})

	return err
}

func (s *Manager) GetMatreshkaKey() string {
	return s.keys.Matreshka
}
func (s *Manager) SetMatreshkaKey(b string) {
	s.keys.Matreshka = b

	err := writeKey(s.buildPath, s.keys)
	if err != nil {
		logrus.Errorf("error setting matreshka key %s", err)
	}
}

func (s *Manager) GetHeadscaleKey() string {
	return s.keys.Headscale
}
func (s *Manager) SetHeadscaleKey(key string) {
	s.keys.Headscale = key

	err := writeKey(s.buildPath, s.keys)
	if err != nil {
		logrus.Errorf("error setting headscale key %s", err)
	}
}

func (s *Manager) ValidateVelezPrivateKey(in string) bool {
	if len(in) != len(s.keys.Velez) {
		return false
	}

	for i := range in {
		if in[i] != s.keys.Velez[i] {
			return false
		}
	}

	return true
}

func (s *Manager) Stop() error {
	return nil
}

func (s *Manager) start() error {
	keys, err := getKeys(s.buildPath)
	if err != nil {
		return rerrors.Wrap(err, "unable to get private keys")
	}

	s.keys.Velez = firstNotEmptyKey(keys.Velez)
	s.keys.Matreshka = firstNotEmptyKey(s.keys.Matreshka, keys.Matreshka)

	err = writeKey(s.buildPath, s.keys)
	if err != nil {
		return rerrors.Wrap(err, "unable to write key")
	}

	return nil
}

func getKeys(buildPath string) (keys PrivateKeys, err error) {
	f, err := os.Open(buildPath)
	if err != nil {
		if os.IsNotExist(err) {
			return keys, nil
		}

		return keys, rerrors.Wrap(err, "error opening file")
	}
	defer f.Close()

	_ = json.NewDecoder(f).Decode(&keys)
	return keys, nil
}

func writeKey(buildPath string, keys PrivateKeys) error {
	err := os.MkdirAll(path.Dir(buildPath), 0777)
	if err != nil {
		return rerrors.Wrap(err, "error making dir")
	}

	data, err := json.Marshal(keys)
	if err != nil {
		return rerrors.Wrap(err, "error encoding key")
	}

	b := bytes.NewBuffer([]byte{})
	err = json.Indent(b, data, "", "  ")
	if err != nil {
		return rerrors.Wrap(err, "error indenting key")
	}
	data = b.Bytes()

	err = os.WriteFile(buildPath, data, 0777)
	if err != nil {
		return rerrors.Wrap(err, "error creating file")
	}

	return nil
}

func firstNotEmptyKey(arrs ...string) string {
	for _, b := range arrs {
		if len(b) > 0 {
			return b
		}
	}

	return string(rtb.RandomBase64(256))
}
