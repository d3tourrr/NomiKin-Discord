package main

import (
    "github.com/bwmarrin/discordgo"
)

var Version = "v0.8.1"
var RoomPrimaries = make(map[string]string)
var Verbose = false
var Companions = make(map[*discordgo.Session]*Companion)

