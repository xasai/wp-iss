package main

type Resp struct {
	Status
	Url string
}

type Status int8

const (
	FAIL Status = iota
	SETUP
	INSTALL
)
