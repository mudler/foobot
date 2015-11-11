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
	Bugs           string
	HalBrainFile   string
	HalEnabled     bool
	HalMarkovOrder int
	Debug          bool
	Welcome        bool
	WelcomeMessage string
	JoinMessage    string
	MessageOnJoin  bool
}
