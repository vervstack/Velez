package common

import (
	"github.com/rs/zerolog/log"
	"go.redsock.ru/toolbox/closer"
)

func CloseWithLog(closable closer.Closable, target string) {
	e := closable()
	if e != nil {
		log.Err(e).
			Str("target", target).
			Msg("error closing closable")
	}
}
