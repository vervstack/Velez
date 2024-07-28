package security

import (
	"os"
	"path"
	"sync"

	rtb "github.com/Red-Sock/toolbox"
	errors "github.com/Red-Sock/trace-errors"
	"github.com/sirupsen/logrus"
)

const defaultPath = "/tmp/velez/private.key"

type Manager struct {
	buildPath string
	key       []byte

	sync.Once
}

func NewSecurityManager(buildPath string) *Manager {
	if buildPath == "" {
		buildPath = defaultPath
	}
	return &Manager{
		buildPath: buildPath,
	}
}

func (s *Manager) Start() error {
	var err error
	s.Once.Do(func() {
		err = s.start()
	})

	return err
}
func (s *Manager) ValidateKey(in string) bool {
	if len(in) != len(s.key) {
		return false
	}

	for i := range in {
		if in[i] != s.key[i] {
			return false
		}
	}

	return true
}

func (s *Manager) Stop() error {
	return os.RemoveAll(s.buildPath)
}

func (s *Manager) start() (err error) {
	s.key, err = rtb.Random(256)
	if err != nil {
		return errors.Wrap(err, "error generating random key")
	}

	logrus.Infof("making key to %s", s.buildPath)

	err = os.RemoveAll(s.buildPath)
	if err != nil {
		return errors.Wrap(err, "error removing old key")
	}

	err = os.MkdirAll(path.Dir(s.buildPath), os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "error making dir")
	}

	err = os.WriteFile(s.buildPath, s.key, 0666)
	if err != nil {
		return errors.Wrap(err, "error writing key")
	}

	logrus.Infof("wrote key to %s", s.buildPath)

	return nil
}
