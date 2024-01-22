package security

import (
	"crypto/rand"
	"os"
	"path"
	"sync"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/sirupsen/logrus"
)

const defaultPath = "/tmp/velez/private.key"

type Validator interface {
	ValidateKey(in string) bool
}

type Manager interface {
	Start() error
	Stop() error

	Validator
}

type manager struct {
	buildPath string
	key       []byte

	sync.Once
}

func NewSecurityManager(buildPath string) Manager {
	if buildPath == "" {
		buildPath = defaultPath
	}
	return &manager{
		buildPath: buildPath,
	}
}

func (s *manager) Start() error {
	var err error
	s.Once.Do(func() {
		err = s.start()
	})

	return err
}
func (s *manager) ValidateKey(in string) bool {
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

func (s *manager) Stop() error {
	return os.RemoveAll(s.buildPath)
}

func (s *manager) start() error {
	s.key = make([]byte, 256)
	_, err := rand.Read(s.key)
	if err != nil {
		return err
	}

	for i := range s.key {
		if s.key[i] > 126 {
			s.key[i] %= 126
		}

		if s.key[i] < 33 {
			s.key[i] += 33
		}
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
