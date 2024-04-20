package closer

import (
	"errors"
	"sync"
)

var m sync.Mutex

type Closable func() error

var funcs []Closable

func Add(f Closable) {
	m.Lock()
	funcs = append(funcs, f)
	m.Unlock()
}

func Close() (err error) {
	for _, f := range funcs {
		fErr := f()
		if fErr != nil {
			err = errors.Join(err, fErr)
		}
	}

	return err
}
