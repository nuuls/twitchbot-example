package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"strings"
)

func main() {
	channel := flag.String("channel", "nuuls", "twitch channel to join")
	username := flag.String("username", "justinfan123", "twitch username")
	oauth := flag.String("oauth", "", "your oauth token")
	flag.Parse()
	chat := connect(*username, *oauth)
	chat.join(*channel)
	for message := range chat.messages {
		fmt.Println(message)
		if strings.HasPrefix(message, "PING") {
			chat.send("PONG :tmi.twitch.tv")
			continue
		}
		// do something with the message here xd
	}
}

type irc struct {
	conn     net.Conn
	messages chan string
}

func connect(username, oauth string) *irc {
	conn, err := tls.Dial("tcp", "irc.chat.twitch.tv:443", nil)
	if err != nil {
		log.Fatal("cannot connect to twitch irc server", err)
	}
	i := &irc{
		conn:     conn,
		messages: make(chan string, 10),
	}
	i.send("CAP REQ twitch.tv/commands")
	i.send("CAP REQ twitch.tv/tags")
	if oauth != "" {
		i.send("PASS " + oauth)
	}
	i.send("NICK " + username)
	go i.read()
	fmt.Println("connected to twitch irc server")
	return i
}

func (i *irc) join(channel string) {
	i.send("JOIN #" + strings.ToLower(channel))
	fmt.Println("joined channel", channel)
}

func (i *irc) send(msg string) {
	_, err := i.conn.Write([]byte(msg + "\r\n"))
	if err != nil {
		log.Fatal("disconnected from twitch irc server", err)
	}
}

func (i *irc) say(channel, msg string) {
	fmt.Printf("sending #%s : %s\n", channel, msg)
	i.send(fmt.Sprintf("PRIVMSG #%s :%s", channel, msg))
}

func (i *irc) read() {
	reader := bufio.NewReader(i.conn)
	tp := textproto.NewReader(reader)
	for {
		message, err := tp.ReadLine()
		if err != nil {
			log.Fatal("disconnected from twitch irc server", err)
		}
		i.messages <- message
	}
}
