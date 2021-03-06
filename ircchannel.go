package hbot

import (
	"os"
	"encoding/json"
	"fmt"
)

//User Mode Flags
const (
	UMAway = uint32(1) << iota
	UMInvisibility
	UMWallops
	UMRestricted
	UMOperator
	UMLocalOp
	UMRecNotices
)

var UserModeMap = map[string]uint32{
	"a" : UMAway,
	"i" : UMInvisibility,
	"w" : UMWallops,
	"r" : UMRestricted,
	"o" : UMOperator,
	"O" : UMLocalOp,
	"s" : UMRecNotices,
}

// Currently Unused
type IrcUser struct {
	Nick string
	User string
	Host string

	Perms uint32
}

type IrcChannel struct {
	Name string
	con *IrcCon
	Counts map[string]int
	Perms uint32


	istream chan *Message
	//ostream chan *Message

	// Currently Unused
	Users map[string]*IrcUser
}

// Attempt to load chat frequency stats from a file
func (c *IrcChannel) TryLoadStats(finame string) bool {
	fi,err := os.Open(finame)
	if err != nil {
		return false
	}
	defer fi.Close()

	dec := json.NewDecoder(fi)

	err = dec.Decode(&c.Counts)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// Write chat frequencies to a file
func (c *IrcChannel) SaveStats(finame string) {
	fi,err := os.Create(finame)
	if err != nil {
		panic(err)
	}
	defer fi.Close()

	enc := json.NewEncoder(fi)
	enc.Encode(c.Counts)
}

//Take note of joins, parts, and mode changes
func (c *IrcChannel) handleMessages() {
	for mes := range c.istream {
		switch mes.Command {
		case "JOIN":
			u := new(IrcUser)
			u.Host = mes.Host
			u.Nick = mes.Name
			u.User = mes.User
			c.Users[mes.Name] = u
		case "MODE":
			ch := mes.Params[1][0]
			u := c.Users[mes.Params[2]]
			if ch == '+' {
				u.Perms |= UserModeMap[mes.Params[1][1:]]
			} else if ch == '-' {
				u.Perms &= ^UserModeMap[mes.Params[1][1:]]
			}
		case "PART":
			delete(c.Users, mes.Name)
		case "NICK":
			newnick := mes.Content
			u := c.Users[mes.From]
			delete(c.Users, mes.From)
			c.Users[newnick] = u
			u.Nick = newnick
		}
	}
}

// Send a message to this irc channel
func (c *IrcChannel) Say(text string) {
	if c == nil {
		fmt.Println("tried to send to channel youre not in...")
		return
	}
	c.con.Send(fmt.Sprintf("PRIVMSG %s :%s\r\n", c.Name, text))
}

// Sets the channels topic (requires bot has proper permissions)
func (c *IrcChannel) Topic(topic string) {
	str := fmt.Sprintf("TOPIC %s :%s", c.Name, topic)
	c.con.Send(str)
}

// Kick a user in this channel, reason optional (requires permissions)
func (c *IrcChannel) Kick(user, reason string) {
	c.con.Send(fmt.Sprintf("KICK %s %s :%s", c.Name, user, reason))
}

/* Commented out until i have a clever way of making it threadsafe

// Returns a channel that will contain messages sent to 
// the channel represented by this IrcChannel
func (c *IrcChannel) GetMessageStream() chan *Message {
	if c.ostream == nil {
		c.ostream = make(chan *Message, 16)
	}
	return c.ostream
}

// Closes out the message stream for this channel
func (c *IrcChannel) CloseMessageStream() {
	close(c.ostream)
	c.ostream = nil
}

*/
