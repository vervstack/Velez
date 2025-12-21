package state

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
	state     State

	sync.Once
	m sync.RWMutex
}

type State struct {
	VelezKey     string
	MatreshkaKey string
	HeadscaleKey string
	PgRootDsn    string
	PgNodeDsn    string
}

func NewSecurityManager(cfg config.Config) *Manager {
	if cfg.Environment.CustomPassToKey == "" {
		cfg.Environment.CustomPassToKey = defaultPath
	}

	return &Manager{
		buildPath: cfg.Environment.CustomPassToKey,
		state:     State{},
	}
}

func (s *Manager) Start() error {
	var err error
	s.Once.Do(func() {
		err = s.start()
	})

	return err
}

func (s *Manager) Set(state State) {
	s.m.Lock()
	s.state = state

	err := writeKey(s.buildPath, s.state)
	s.m.Unlock()
	if err != nil {
		logrus.Errorf("error setting matreshka key %s", err)
	}
}

func (s *Manager) Get() State {
	s.m.RLock()
	state := s.state
	s.m.RUnlock()
	return state
}

func (s *Manager) GetForUpdate() State {
	s.m.Lock()
	return s.state
}
func (s *Manager) SetAndRelease(state State) {
	s.state = state
	s.m.Unlock()

	s.Set(s.state)
}

func (s *Manager) ValidateVelezPrivateKey(in string) bool {
	if len(in) != len(s.state.VelezKey) {
		return false
	}

	for i := range in {
		if in[i] != s.state.VelezKey[i] {
			return false
		}
	}

	return true
}

func (s *Manager) Stop() error {
	err := writeKey(s.buildPath, s.state)
	if err != nil {
		return rerrors.Wrap(err, "error saving state on exit")
	}

	return nil
}

func (s *Manager) start() error {
	keys, err := readStateFromPath(s.buildPath)
	if err != nil {
		return rerrors.Wrap(err, "unable to get private keys")
	}
	s.state = keys

	s.state.VelezKey = firstNotEmptyKey(keys.VelezKey)
	s.state.MatreshkaKey = firstNotEmptyKey(s.state.MatreshkaKey, keys.MatreshkaKey)

	err = writeKey(s.buildPath, s.state)
	if err != nil {
		return rerrors.Wrap(err, "unable to write key")
	}

	return nil
}

func readStateFromPath(buildPath string) (state State, err error) {
	f, err := os.Open(buildPath)
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil
		}

		return state, rerrors.Wrap(err, "error opening file")
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&state)
	if err != nil {
		return state, rerrors.Wrap(err, "error decoding local state file")
	}
	return state, nil
}

func writeKey(buildPath string, keys State) error {
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
