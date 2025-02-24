package main

import (
    NomiKin "github.com/d3tourrr/NomiKinGo"
)

type NomiRoom struct {
    Name    string
    Note    string
    Uuid    string
    Backchanneling bool
    Nomis   []NomiKin.Nomi
    RandomResponseChance int
}

type KinRoom struct {
    ID      string
    RandomResponseChance int
}
