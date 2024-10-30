package main

import (
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/app"
)

func main() {
	a, err := app.New()
	if err != nil {
		logrus.Fatal(err)
	}

	a.Start()
}
