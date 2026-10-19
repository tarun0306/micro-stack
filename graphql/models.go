package main

type Account struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Role   string   `json:"role"`
	Orders []*Order `json:"orders"`
}
