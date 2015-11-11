/*
GoBot

An IRC bot written in Go.

Copyright (C) 2014  Brian C. Tomlinson
Copyright (C) 2015  Ettore Di Giacinto

Contact: mudler@sabayon.org

This program is free software; you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation; either version 2 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License along
with this program; if not, write to the Free Software Foundation, Inc.,
51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"

	"github.com/thoj/go-ircevent"
)

const delay = 40

func main() {
	configfile := "config.json"
	if os.Args[1] != "" {
		configfile = os.Args[1]
		fmt.Println("Reading " + configfile)
	}

	rand.Seed(64)
	file, err := os.Open(configfile)

	if err != nil {
		fmt.Println("Couldn't read config file, dying...")
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := &Config{}
	config.HalEnabled = false
	config.HalMarkovOrder = 4
	config.MessageOnJoin = false
	config.Welcome = false
	config.Debug = false
	decoder.Decode(&config)

	conn := irc.IRC(config.BotNick, config.BotUser)
	err = conn.Connect(config.Server)
	if config.Debug {
		conn.Debug = true
	}
	fmt.Println("BotNick:\t" + config.BotNick)
	fmt.Println("BotUser:\t" + config.BotUser)
	if config.HalEnabled {
		fmt.Println("Hal:\tenabled")
	} else {
		fmt.Println("Hal:\tdisabled")
	}
	fmt.Println("Channels:")
	for _, channel := range config.Channel {
		fmt.Println("\t" + channel)
	}
	fmt.Println("Admins:")
	for _, admin := range config.Admins {
		fmt.Println("\t" + admin)
	}
	if err != nil {
		fmt.Println("Failed to connect.")
		panic(err)
	}

	AddCallbacks(conn, config)
	conn.Loop()
}
