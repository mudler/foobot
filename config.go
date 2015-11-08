package main

type Config struct {
	Admins         []string
	Server         string
	Channel        []string
	BotUser        string
	BotNick        string
	Trigger        string
	WeatherKey     string
	LogDir         string
	WikiLink       string
	Homepage       string
	Forums         string
	HalBrainFile   string
	HalEnabled     bool
	HalMarkovOrder int
	Debug          bool
}
