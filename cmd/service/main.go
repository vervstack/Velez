package main

import (
	"github.com/sirupsen/logrus"
	"go.redsock.ru/toolbox/respect"

	"github.com/godverv/Velez/internal/app"
)

func main() {
	println(respect.Fantasy)

	a, err := app.New()
	if err != nil {
		logrus.Fatal(err)
	}

	err = a.Start()
	if err != nil {
		logrus.Fatal(err)
	}
}
