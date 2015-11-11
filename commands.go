package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/darthlukan/cakeday"
	"github.com/darthlukan/goconvtemps"
	"github.com/darthlukan/goduckgo/goduckgo"
	"github.com/freahs/microhal"
	"github.com/thoj/go-ircevent"
)

// ParseCmds takes PRIVMSG strings containing a preceding bang "!"
// and attempts to turn them into an ACTION that makes sense.
// Returns a msg string.
func ParseCmds(cmdMsg string, config *Config, conn *irc.Connection, e *irc.Event) string {
	var (
		msg      string
		msgArray []string
		cmdArray []string
	)

	cmdArray = strings.SplitAfterN(cmdMsg, config.Trigger, 2)

	if len(cmdArray) > 0 {
		msgArray = strings.SplitN(cmdArray[1], " ", 2)
	}

	if len(msgArray) > 1 {
		cmd := fmt.Sprintf("%vs", msgArray[0])
		switch {
		case strings.Contains(cmd, "cakeday"):
			msg = CakeDayCmd(msgArray[1])
		case strings.Contains(cmd, "ddg"), strings.Contains(cmd, "search"):
			query := strings.Join(msgArray[1:], " ")
			msg = SearchCmd(query)
		case strings.Contains(cmd, "convtemp"):
			query := strings.Join(msgArray[1:], " ")
			msg = ConvertTempCmd(query)
		case strings.Contains(cmd, "rand"):
			msg = GenericVerbCmd(cmd, msgArray[1])
		case strings.Contains(cmd, "pkg"):
			go SearchPkgsCmd(conn, e, strings.Join(msgArray[1:], " "), "SearchPackage")
			msg = ""
		case strings.Contains(cmd, "rdep"):
			go SearchPkgsCmd(conn, e, strings.Join(msgArray[1:], " "), "SearchRevDeps")
			msg = ""
		default:
			msg = GenericVerbCmd(cmd, msgArray[1])
		}
	} else {
		switch {
		case strings.Contains(msgArray[0], "help"):
			HelpCmd(conn, e, config.Trigger)
			msg = ""
		case strings.Contains(msgArray[0], "wiki"):
			msg = WikiCmd(config)
		case strings.Contains(msgArray[0], "homepage"):
			msg = HomePageCmd(config)
		case strings.Contains(msgArray[0], "forum"):
			msg = ForumCmd(config)
		case strings.Contains(msgArray[0], "bugs"):
			msg = BugsCmd(config)
		case strings.Contains(msgArray[0], "latestpkgs"):
			go SearchPkgsCmd(conn, e, "", "SearchPackage")
			msg = ""
		default:
			msg = ""
		}
	}
	return msg
}

// GenericVerbCmd returns a message string based on the supplied cmd (a verb).
func GenericVerbCmd(cmd, extra string) string {
	randQuip := RandomQuip()
	return fmt.Sprintf("\x01"+"ACTION %v %v, %v\x01", cmd, extra, randQuip)
}

// CakeDayCmd returns a string containing the Reddit cakeday of a user
// upon success, or an error string on failure.
func CakeDayCmd(user string) string {
	var msg string

	responseString, err := cakeday.Get(user)
	if err != nil {
		msg = fmt.Sprintf("I caught an error: %v\n", err)
	} else {
		msg = fmt.Sprintf("%v\n", responseString)
	}
	return msg
}

// WebSearch takes a query string as an argument and returns
// a formatted string containing the results from DuckDuckGo.
func SearchCmd(query string) string {
	msg, err := goduckgo.Query(query)
	if err != nil {
		return fmt.Sprintf("DDG Error: %v\n", err)
	}

	switch {
	case len(msg.RelatedTopics) > 0:
		return fmt.Sprintf("First Topical Result: [ %s ]( %s )\n", msg.RelatedTopics[0].FirstURL, msg.RelatedTopics[0].Text)
	case len(msg.Results) > 0:
		return fmt.Sprintf("First External result: [ %s ]( %s )\n", msg.Results[0].FirstURL, msg.Results[0].Text)
	case len(msg.Redirect) > 0:
		return fmt.Sprintf("Redirect result: %s\n", UrlTitle(msg.Redirect))
	default:
		return fmt.Sprintf("Query: '%s' returned no results.\n", query)
	}
}

func ConvertTempCmd(query string) string {
	var unit string
	var converted string

	input := strings.ToLower(query)

	if strings.Index(input, "c") != -1 {
		unit = "c"
	} else if strings.Index(input, "f") != -1 {
		unit = "f"
	} else {
		return fmt.Sprintf("Invalid unit input, please use either 'F' or 'C'.\n")
	}

	temp, err := strconv.ParseFloat(fmt.Sprintf("%v", string(strings.Split(input, unit)[0])), 64)

	if err != nil {
		return fmt.Sprintf("Caught error '%v' trying to convert '%v'.\n", err, query)
	}

	converted = goconvtemps.ConvertTemps(temp, unit)

	return fmt.Sprintf("%v is %v.\n", strings.ToUpper(input), converted)
}

func HelpCmd(conn *irc.Connection, e *irc.Event, trigger string) {
	conn.Privmsg(e.Arguments[0], "Available commands:")
	conn.Privmsg(e.Arguments[0], "General info: "+trigger+"forum,  "+trigger+"homepage,  "+trigger+"wiki,  "+trigger+"bugs")
	conn.Privmsg(e.Arguments[0], "Sabayon Entropy store search: "+trigger+"latestpkgs (show you latest packages), "+trigger+"pkg <package> (search for a package), "+trigger+"rdep <package> (reverse dependency of a package)")
	conn.Privmsg(e.Arguments[0], "Various utils: "+trigger+"ddg/search <whatever>, "+trigger+"convtemp <27C>, "+trigger+"cakeday <someone>, "+trigger+"random <whatever>, ")
}

func WikiCmd(config *Config) string {
	return fmt.Sprintf("(Wiki)[ %s ]\n", config.WikiLink)
}

func HomePageCmd(config *Config) string {
	return fmt.Sprintf("(Homepage)[ %s ]\n", config.Homepage)
}

func ForumCmd(config *Config) string {
	return fmt.Sprintf("(Forums)[ %s ]\n", config.Forums)
}
func BugsCmd(config *Config) string {
	return fmt.Sprintf("(Bugs)[ %s ]\n", config.Bugs)
}

func QuitCmd(admins []string, user string) {
	for _, admin := range admins {
		if user == admin {
			os.Exit(0)
		}
	}
}

var quips = []string{
	"FOR SCIENCE!",
	"because... reasons.",
	"it's super effective!",
	"because... why not?",
	"was it good for you?",
	"given the alternative, yep, worth it!",
	"don't ask...",
	"then makes a sandwich.",
	"oh noes!",
	"did I do that?",
	"why must you turn this place into a house of lies!",
	"really???",
	"LLLLEEEEEERRRRRROOOOYYYY JEEEENNNKINNNS!",
	"DOH!",
	"Giggity!",
}

func SearchPkgsCmd(conn *irc.Connection, e *irc.Event, s string, t string) {
	max := 3
	var search []Package
	var query string
	conn.Privmsg(e.Arguments[0], "Searching, be patient boy")
	if t == "SearchPackage" {
		if len(s) >= 2 {
			search, query = SearchPackages(s)
		} else {
			search, query = SearchPackages("")
		}
	} else if t == "SearchRevDeps" {
		if len(s) >= 2 {
			search, query = ReverseDeps(s)
		} else {
			conn.Privmsg(e.Arguments[0], "Please, be more specific next time")
			return
		}
	}
	if len(search) == 0 {
		conn.Privmsg(e.Arguments[0], "No results for "+query+" limited to "+strconv.Itoa(max)+" results")

	} else {
		conn.Privmsg(e.Arguments[0], "Showing the results limited to "+strconv.Itoa(max)+" for "+query)
	}
	if len(search) < max {
		max = len(search)
	}
	for i := 0; i < max; i++ {
		conn.Privmsg(e.Arguments[0], search[i].String())
		time.Sleep(1000 * time.Millisecond)
	}
}

func RandomQuip() string {
	return quips[rand.Intn(len(quips))]
}

// UrlTitle attempts to extract the title of the page that a
// pasted URL points to.
// Returns a string message with the title and URL on success, or a string
// with an error message on failure.
func UrlTitle(msg string) string {
	var (
		newMsg, url, title, word string
	)

	regex, _ := regexp.Compile(`(?i)<title>(.*?)<\/title>`)

	msgArray := strings.Split(msg, " ")

	for _, word = range msgArray {
		if strings.Contains(word, "http") {
			url = word
			break
		}
		if !strings.Contains(word, "http") && strings.Contains(word, "www") {
			url = "http://" + word
			break
		}
	}

	resp, err := http.Get(url)

	if err != nil {
		return fmt.Sprintf("Could not resolve URL %v, beware...\n", url)
	}

	defer resp.Body.Close()

	rawBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Could not read response Body of %v ...\n", url)
	}

	body := string(rawBody)
	noNewLines := strings.Replace(body, "\n", "", -1)
	noCarriageReturns := strings.Replace(noNewLines, "\r", "", -1)
	notSoRawBody := noCarriageReturns

	titleMatch := regex.FindStringSubmatch(notSoRawBody)
	if len(titleMatch) > 1 {
		title = strings.TrimSpace(titleMatch[1])
	} else {
		title = fmt.Sprintf("Title Resolution Failure")
	}
	newMsg = fmt.Sprintf("[ %v ]( %v )\n", title, url)

	return newMsg
}

// AddCallbacks is a single function that does what it says.
// It's merely a way of decluttering the main function.
func AddCallbacks(conn *irc.Connection, config *Config) {

	conn.AddCallback("001", func(e *irc.Event) {
		for _, channel := range config.Channel {
			conn.Join(channel)
		}
	})

	if config.Welcome {

		conn.AddCallback("JOIN", func(e *irc.Event) {
			conn.Privmsg(e.Arguments[0], config.WelcomeMessage)
		})
	}

	conn.AddCallback("JOIN", func(e *irc.Event) {
		if e.Nick == config.BotNick {
			if config.MessageOnJoin {
				conn.Privmsg(e.Arguments[0], config.JoinMessage)
			}
			LogDir(config.LogDir)
			LogFile(config.LogDir + e.Arguments[0])
		}
		message := fmt.Sprintf("%s has joined", e.Nick)
		go ChannelLogger(config.LogDir+e.Arguments[0], e.Nick, message)
	})
	conn.AddCallback("PART", func(e *irc.Event) {
		message := fmt.Sprintf("has parted (%s)", e.Message())
		nick := fmt.Sprintf("%s@%s", e.Nick, e.Host)
		go ChannelLogger(config.LogDir+e.Arguments[0], nick, message)
	})
	conn.AddCallback("QUIT", func(e *irc.Event) {
		message := fmt.Sprintf("has quit (%v)", e.Message)
		nick := fmt.Sprintf("%s@%s", e.Nick, e.Host)
		go ChannelLogger(config.LogDir+e.Arguments[0], nick, message)
	})

	if config.HalEnabled {
		var brain *microhal.Microhal

		if _, err := os.Stat(config.HalBrainFile + ".json"); os.IsNotExist(err) {
			brain = microhal.NewMicrohal(config.HalBrainFile, config.HalMarkovOrder)
		} else {
			brain = microhal.LoadMicrohal(config.HalBrainFile)
		}

		re, _ := regexp.Compile(config.BotNick)
		brainIn, brainOut := brain.Start(10000*time.Millisecond, 250)
		conn.AddCallback("PRIVMSG", func(e *irc.Event) {
			message := e.Message()
			sanitizedInput := re.ReplaceAllLiteralString(message, "")
			if !strings.HasPrefix(message, config.Trigger) && len(message) >= config.HalMarkovOrder {
				brainIn <- sanitizedInput
				res := <-brainOut
				if sanitizedInput != message {
					conn.Privmsg(e.Arguments[0], res)
				}
			} else if len(message) >= config.HalMarkovOrder {
				brainIn <- sanitizedInput
				_ = <-brainOut
			}
		})

	}

	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		var response string
		message := e.Message()
		if strings.Contains(message, config.Trigger) && strings.Index(message, config.Trigger) == 0 {
			response = ParseCmds(message, config, conn, e)
		}
		if strings.Contains(message, "http://") || strings.Contains(message, "https://") || strings.Contains(message, "www.") {
			response = UrlTitle(message)
		}

		if strings.Contains(message, fmt.Sprintf("%squit", config.Trigger)) {
			QuitCmd(config.Admins, e.Nick)
		}

		if len(response) > 0 {
			conn.Privmsg(e.Arguments[0], response)
		}

		if len(message) > 0 {
			if e.Arguments[0] != config.BotNick {
				go ChannelLogger(config.LogDir+e.Arguments[0], e.Nick+": ", message)
			}
		}
	})
}
