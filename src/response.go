package main

type Resp struct {
	Status
	Url string
}

type Status int8

const (
	_ Status = iota
	SETUP
	FAIL
	INSTALL
)
