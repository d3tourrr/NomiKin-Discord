package main

import (
    "github.com/bwmarrin/discordgo"
)

var Version = "v0.8.2"
var RoomPrimaries = make(map[string]string)
var Verbose = false
var LogWidth = 5
var Companions = make(map[*discordgo.Session]*Companion)
var Banner = `
   _  __           _ __ ___               
  / |/ /__  __ _  (_) //_(_)__            
 /    / _ \/  ' \/ / ,< / / _ \           
/_/|_/\___/_/_/_/_/_/|_/_/_//_/     __    
    ____/ _ \(_)__ _______  _______/ /    
   /___/ // / (_-</ __/ _ \/ __/ _  /     
   __ /____/_/___/\__/\___/_/  \_,_/      
  / /  __ __  ___/ /_  // /____  __ ______
 / _ \/ // / / _  //_ </ __/ _ \/ // / __/
/_.__/\_, /  \_,_/____/\__/\___/\_,_/_/   
     /___/                                
`

