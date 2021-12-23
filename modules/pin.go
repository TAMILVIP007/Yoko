package modules

import (
	"fmt"

	tb "gopkg.in/tucnak/telebot.v3"
)

func pin_message(c tb.Context) error {
	args := c.Message().Payload
	if c.Message().ReplyTo == nil {
		c.Reply("You need to reply to a message to pin it!")
		return nil
	}
	silent := false
	if args != string("") {
		sup := false
		if stringInSlice(args, []string{"loud", "violent", "notify"}) {
			sup = true
			silent = false
		} else if stringInSlice(args, []string{"quiet", "silent"}) {
			sup = true
			silent = true
		}
		if !sup {
			c.Reply(fmt.Sprintf("'%s' was not recognised as a valid pin option. Please use one of: <b><i>loud/violent/notify/silent/quiet</i></b>", args))
			return nil
		}
	}
	if !silent {
		c.Bot().Pin(c.Message().ReplyTo)
	} else {
		c.Bot().Pin(c.Message().ReplyTo, tb.Silent)
	}
	c.Reply(fmt.Sprintf("I have pinned <a href='t.me/c/%d/%d'>this message</a>", c.Message().ReplyTo.Chat.ID, c.Message().ReplyTo.ID))
	return nil
}