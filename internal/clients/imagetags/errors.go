package imagetags

import (
	"go.redsock.ru/rerrors"
)

var errHubUnexpectedStatus = rerrors.New("docker hub returned unexpected status")
