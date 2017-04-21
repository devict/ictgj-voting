package main

import "github.com/pborman/uuid"

type Game struct {
	UUID *uuid.UUID
	Name string
}
