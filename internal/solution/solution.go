package solution

import (
	"fmt"
)

type Project struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Target []string `json:"target"`
	Module []string `json:"module,omitempty"`
}

// Sln is JSON struct
type Sln struct {
	Name    string    `json:"name"`
	Lang    string    `json:"lang"`
	Project []Project `json:"project,omitempty"`
}

type Info struct {
	Deletion int
	Addition int
}

func (i Info) Delete() {
	i.Deletion++
}
func (i Info) Add() {
	i.Addition++
}

func (i Info) Show() {
	fmt.Println(i.Addition, `files added`)
	fmt.Println(i.Deletion, `files deleted`)
}
