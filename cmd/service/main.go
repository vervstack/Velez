package main

import (
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/app"
)

func main() {
	err := app.New().Start()
	if err != nil {
		logrus.Fatal(err)
	}
}
