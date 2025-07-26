package security

import (
	"encoding/json"
	"os"
	"path"
	"sync"

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
	Velez     []byte
	Matreshka []byte
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

func (s *Manager) PrivateKeys() *PrivateKeys {
	return &s.keys
}
func (s *Manager) ValidatePrivateKey(in string) bool {
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
	return os.RemoveAll(s.buildPath)
}

func (s *Manager) start() error {
	keys, err := getKeys(s.buildPath)
	if err != nil {
		return rerrors.Wrap(err, "unable to get private keys")
	}

	s.keys.Velez = firstNotEmptyKey(keys.Velez)
	s.keys.Matreshka = firstNotEmptyKey(s.keys.Velez, keys.Matreshka)

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

	err = os.WriteFile(buildPath, data, 0777)
	if err != nil {
		return rerrors.Wrap(err, "error creating file")
	}

	return nil
}

func firstNotEmptyKey(arrs ...[]byte) []byte {
	for _, b := range arrs {
		if len(b) > 0 {
			return b
		}
	}

	return rtb.RandomBase64(256)
}
