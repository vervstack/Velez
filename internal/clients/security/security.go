package security

import (
	"os"
	"path"
	"sync"

	"github.com/sirupsen/logrus"
	errors "go.redsock.ru/rerrors"
	rtb "go.redsock.ru/toolbox"
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
	s.key = rtb.RandomBase64(256)

	logrus.Debugf("making key to %s", s.buildPath)

	err = os.RemoveAll(s.buildPath)
	if err != nil {
		return errors.Wrap(err, "error removing old key")
	}

	err = os.MkdirAll(path.Dir(s.buildPath), 0777)
	if err != nil {
		return errors.Wrap(err, "error making dir")
	}

	err = os.WriteFile(s.buildPath, s.key, 0777)
	if err != nil {
		return errors.Wrap(err, "error writing key")
	}

	logrus.Infof("Private keys are at %s", s.buildPath)

	return nil
}
