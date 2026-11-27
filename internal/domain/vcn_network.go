package domain

import (
	"time"
)

type Network struct {
	Id          string
	Name        string
	Provider    string
	Source      string
	Url         string
	Token       string
	ServiceName string
	ContainerId string
	NodeName    string
	CreatedAt   time.Time
}
