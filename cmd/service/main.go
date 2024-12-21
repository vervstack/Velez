package main

import (
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/app"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	a, err := app.New()
	if err != nil {
		logrus.Fatal(err)
	}

	err = a.Start()
	if err != nil {
		logrus.Fatal(err)
	}
}
