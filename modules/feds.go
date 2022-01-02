package modules

import (
	"fmt"
	"strconv"
	"strings"

	db "github.com/amarnathcjd/yoko/modules/db"
	tb "gopkg.in/tucnak/telebot.v3"
)

var (
	sel               = &tb.ReplyMarkup{}
	accept_fpromote   = sel.Data("Accept", "accept_fpromote")
	deny_fpromote     = sel.Data("Decline", "decline_fpromote")
	accept_ftransfer  = sel.Data("Accept", "accept_ftransfer")
	deny_ftransfer    = sel.Data("Decline", "deny_ftransfer")
	confirm_ftransfer = sel.Data("Confirm", "confirm_ftransfer")
	reject_ftransfer  = sel.Data("Cancel", "reject_ftransfer")
)

func New_fed(c tb.Context) error {
	m := c.Message()
	if !m.Private() {
		c.Reply("Create your federation in my PM - not in a group.")
		return nil
	}
	fed, _, fedname := db.Get_fed_by_owner(m.Sender.ID)
	if fed {
		c.Reply(fmt.Sprintf("You already have a federation called <code>%s</code> ; you can't create another. If you would like to rename it, use <code>/renamefed</code>.", fedname))
		return nil
	}
	if m.Payload == string("") {
		c.Reply("You need to give your federation a name! Federation names can be up to 64 characters long.")
		return nil
	} else if len(m.Payload) > 64 {
		c.Reply("Federation names can only be upto 64 charactors long.")
		return nil
	}
	fed_uid, _ := db.Make_new_fed(m.Sender.ID, m.Payload)
	c.Reply(fmt.Sprintf("Created new federation with FedID: <code>%s</code>.\nUse this ID to join the federation! eg:\n<code>/joinfed %s</code>", fed_uid, fed_uid))
	return nil
}

func Delete_fed(c tb.Context) error {
	m := c.Message()
	if !m.Private() {
		c.Reply("Delete your federation in my PM - not in a group.")
		return nil
	}
	fed, fed_id, fedname := db.Get_fed_by_owner(m.Sender.ID)
	if !fed {
		c.Reply("It doesn't look like you have a federation yet!")
		return nil
	}
	menu := &tb.ReplyMarkup{}
	menu.Inline(menu.Row(menu.Data("Delete Federation", fmt.Sprintf("delfed_%s", fed_id))), menu.Row(menu.Data("Cancel", "cancel_fed_delete")))
	c.Reply(fmt.Sprintf("Are you sure you want to delete your federation? This action cannot be undone - you will lose your entire ban list, and '%s' will be permanently gone.", fedname), menu)
	return nil
}

func Rename_fed(c tb.Context) error {
	m := c.Message()
	if !m.Private() {
		c.Reply("You can only rename your fed in PM.")
		return nil
	}
	fed, fed_id, fedname := db.Get_fed_by_owner(m.Sender.ID)
	if !fed {
		c.Reply("It doesn't look like you have a federation yet!")
		return nil
	} else if m.Payload == string("") {
		c.Reply("You need to give your federation a new name! Federation names can be up to 64 characters long.")
		return nil
	}
	db.Rename_fed_by_id(fed_id, m.Payload)
	c.Reply(fmt.Sprintf("Tada! I've renamed your federation from '%s' to '%s'. (FedID: <code>%s</code>).", fedname, m.Payload, fed_id))
	return nil
}

func Join_fed(c tb.Context) error {
	if c.Message().Private() {
		c.Reply("Only supergroups can join feds.")
		return nil
	}
	p, _ := c.Bot().ChatMemberOf(c.Chat(), c.Sender())
	if p.Role == "administrator" {
		c.Reply("You need to be the chat creator to do this!")
		return nil
	} else if p.Role == tb.Member {
		c.Reply("You need to be an admin to do this!")
		return nil
	} else if p.Role != tb.Creator {
		c.Reply("You need to be chat creator to do this!")
		return nil
	}
	args := c.Message().Payload
	if args == string("") {
		c.Reply("You need to specify which federation you're asking about by giving me a FedID!")
		return nil
	} else if len(args) < 10 {
		c.Reply("This isn't a valid FedID format!")
		return nil
	}
	getfed := db.Search_fed_by_id(args)
	if getfed == nil {
		c.Reply("This FedID does not refer to an existing federation.")
		return nil
	} else {
		c.Reply(fmt.Sprintf("Successfully joined the '%s' federation! All new federation bans will now also remove the members from this chat.", getfed["fedname"]))
		db.Chat_join_fed(args, c.Chat().ID)
	}
	return nil
}

func Leave_fed(c tb.Context) error {
	if c.Message().Private() {
		c.Reply("Only supergroups can join feds.")
		return nil
	}
	p, _ := c.Bot().ChatMemberOf(c.Chat(), c.Sender())
	if p.Role == "administrator" {
		c.Reply("You need to be the chat creator to do this!")
		return nil
	} else if p.Role == tb.Member {
		c.Reply("You need to be an admin to do this!")
		return nil
	} else if p.Role != tb.Creator {
		c.Reply("You need to be chat creator to do this!")
		return nil
	}
	chat_fed := db.Get_chat_fed(c.Chat().ID)
	if chat_fed == string("") {
		c.Reply("This chat isn't currently in any federations!")
	} else {
		fed := db.Search_fed_by_id(chat_fed)
		c.Reply(fmt.Sprintf("Chat %s has left the '%s' federation.", c.Chat().Title, fed["fedname"].(string)))
		db.Chat_leave_fed(c.Chat().ID)
	}
	return nil
}

func Chat_fed(c tb.Context) error {
	if c.Message().Private() {
		c.Reply("This command is for supergroups only!")
		return nil
	}
	p, _ := c.Bot().ChatMemberOf(c.Chat(), c.Sender())
	if p.Role != tb.Creator && p.Role != tb.Administrator {
		c.Reply("You need to be an admin to do this.")
		return nil
	}
	fed_id := db.Get_chat_fed(c.Chat().ID)
	if fed_id == string("") {
		c.Reply("This chat isn't part of any feds yet!")
	} else {
		fd := db.Search_fed_by_id(fed_id)
		c.Reply(fmt.Sprintf("Chat %s is part of the following federation: %s (ID: <code>%s</code>)", c.Chat().Title, fd["fedname"].(string), fed_id))
	}
	return nil
}

func Fpromote(c tb.Context) error {
	if c.Message().Private() {
		c.Reply("This command is made to be used in group chats, not in pm!")
		return nil
	}
	user, _ := get_user(c.Message())
	if user == nil {
		return nil
	}
	fed, fed_id, fedname := db.Get_fed_by_owner(c.Sender().ID)
	if !fed {
		c.Reply("Only federation creators can promote people, and you don't seem to have a federation to promote to!")
		return nil
	} else if user.ID == c.Sender().ID {
		c.Reply("Yeah well you are the fed owner!")
		return nil
	}
	fban, reason := db.Is_Fbanned(user.ID, fed_id)
	if fban {
		if reason != string("") {
			reason = fmt.Sprintf("\nReason: %s", reason)
		}
		c.Reply(fmt.Sprintf("User <a href='tg://user?id=%d'>%s</a> is fbanned in %s. You should unfban them before promoting.%s", user.ID, user.FirstName, fedname, reason))
		return nil
	}
	if db.Is_user_fed_admin(user.ID, fed_id) {
		c.Reply(fmt.Sprintf("<a href='tg://user?id=%d'>%s</a> is already an admin in %s!", user.ID, user.FirstName, fedname))
		return nil
	}
	accept_fpromote.Data = strconv.Itoa(int(c.Sender().ID)) + "|" + strconv.Itoa(int(user.ID))
	deny_fpromote.Data = strconv.Itoa(int(c.Sender().ID)) + "|" + strconv.Itoa(int(user.ID))
	sel.Inline(sel.Row(accept_fpromote, deny_fpromote))
	c.Reply(fmt.Sprintf("Please get <a href='tg://user?id=%d'>%s</a> to confirm that they would like to be fed admin for %s", user.ID, user.FirstName, fedname), sel)
	return nil
}

func Fpromote_cb(c tb.Context) error {
	data := strings.SplitN(c.Callback().Data, "|", 2)
	owner_id_int, _ := strconv.Atoi(data[0])
	user_id_int, _ := strconv.Atoi(data[1])
	owner_id := int64(owner_id_int)
	user_id := int64(user_id_int)
	if c.Sender().ID != user_id {
		c.Respond(&tb.CallbackResponse{Text: "You are not the user being fpromoted", ShowAlert: true})
		return nil
	}
	_, fed_id, fedname := db.Get_fed_by_owner(owner_id)
	db.User_join_fed(fed_id, user_id)
	c.Edit(fmt.Sprintf("User <a href='tg://user?id=%d'>%s</a> is now an admin of %s (<code>%s</code>)", user_id, c.Sender().FirstName, fedname, fed_id))
	return nil
}

func Fpromote_deny_cb(c tb.Context) error {
	data := strings.SplitN(c.Callback().Data, "|", 2)
	owner_id_int, _ := strconv.Atoi(data[0])
	user_id_int, _ := strconv.Atoi(data[1])
	owner_id := int64(owner_id_int)
	user_id := int64(user_id_int)
	if c.Sender().ID == owner_id {
		c.Edit(fmt.Sprintf("Fedadmin promotion cancelled by <a href='tg://user?id=%d'>%s</a>", c.Sender().ID, c.Sender().FirstName))
		return nil
	} else if c.Sender().ID == user_id {
		c.Edit(fmt.Sprintf("Fedadmin promotion has been refused by <a href='tg://user?id=%d'>%s</a>", c.Sender().ID, c.Sender().FirstName))
		return nil
	} else {
		c.Respond(&tb.CallbackResponse{Text: "You are not the user being fpromoted", ShowAlert: true})
		return nil
	}
}

func Fdemote(c tb.Context) error {
	if c.Message().Private() {
		c.Reply("This command is made to be used in group chats, not in pm!")
		return nil
	}
	user, _ := get_user(c.Message())
	if user == nil {
		return nil
	}
	fed, fed_id, fedname := db.Get_fed_by_owner(c.Sender().ID)
	if !fed {
		c.Reply("Only federation creators can demote people, and you don't seem to have a federation to promote to!")
		return nil
	} else if user.ID == c.Sender().ID {
		c.Reply("Yeah well you are the fed owner!")
		return nil
	}
	if !db.Is_user_fed_admin(user.ID, fed_id) {
		c.Reply(fmt.Sprintf("This person isn't a federation admin for '%s', how could I demote them?", fedname))
		return nil
	}
	db.User_leave_fed(fed_id, user.ID)
	c.Reply(fmt.Sprintf("User <a href='tg://user?id=%d'>%s</a> is no longer an admin of %s (<code>%s</code>)", user.ID, user.FirstName, fedname, fed_id))
	return nil
}

func Transfer_fed_user(c tb.Context) error {
	if c.Message().Private() {
		c.Reply("This command is made to be used in group chats, not in pm!")
		return nil
	}
	user, _ := get_user(c.Message())
	if user == nil {
		return nil
	}
	fed, fed_id, fedname := db.Get_fed_by_owner(c.Sender().ID)
	if !fed {
		c.Reply("You don't have a fed to transfer!")
		return nil

	} else if user.ID == c.Sender().ID {
		c.Reply("You can only transfer your fed to others!")
		return nil
	}
	fed_2, _, _ := db.Get_fed_by_owner(user.ID)
	if fed_2 {
		c.Reply(fmt.Sprintf("<a href='tg://user?id=%d'>%s</a> already owns a federation - they can't own another.", user.ID, user.FirstName))
		return nil
	}
	if !db.Is_user_fed_admin(user.ID, fed_id) {
		c.Reply(fmt.Sprintf("<a href='tg://user?id=%d'>%s</a> isn't an admin in %s - you can only give your fed to other admins.", user.ID, user.FirstName, fedname))
		return nil
	}
	accept_ftransfer.Data = strconv.Itoa(int(c.Sender().ID)) + "|" + strconv.Itoa(int(user.ID))
	deny_ftransfer.Data = strconv.Itoa(int(c.Sender().ID)) + "|" + strconv.Itoa(int(user.ID))
	sel.Inline(sel.Row(accept_ftransfer, deny_ftransfer))
	c.Reply(fmt.Sprintf("<a href='tg://user?id=%d'>%s</a>, please confirm you would like to receive fed %s (<code>%s</code>) from <a href='tg://user?id=%d'>%s</a>", user.ID, user.FirstName, fedname, fed_id, c.Sender().ID, c.Sender().FirstName), sel)
	return nil
}

func Accept_Transfer_fed_cb(c tb.Context) error {
	data := strings.SplitN(c.Callback().Data, "|", 2)
	owner_id_int, _ := strconv.Atoi(data[0])
	user_id_int, _ := strconv.Atoi(data[1])
	owner_id := int64(owner_id_int)
	user_id := int64(user_id_int)
	if c.Sender().ID != user_id {
		c.Respond(&tb.CallbackResponse{Text: "This action is not intended for you.", ShowAlert: true})
		return nil
	}
	_, fed_id, fedname := db.Get_fed_by_owner(owner_id)
	owner, _ := c.Bot().ChatByID(owner_id)
	confirm_ftransfer.Data = strconv.Itoa(int(owner_id)) + "|" + strconv.Itoa(int(c.Sender().ID))
	reject_ftransfer.Data = strconv.Itoa(int(owner_id)) + "|" + strconv.Itoa(int(c.Sender().ID))
	sel.Inline(sel.Row(confirm_ftransfer, reject_ftransfer))
	c.Edit(fmt.Sprintf("<a href='tg://user?id=%d'>%s</a>, please confirm that you wish to send fed %s (<code>%s</code>) to <a href='tg://user?id=%d'>%s</a> this cannot be undone.", owner_id, owner.FirstName, fedname, fed_id, c.Sender().ID, c.Sender().FirstName), sel)
	return nil
}

func Decline_Transfer_fed_cb(c tb.Context) error {
	data := strings.SplitN(c.Callback().Data, "|", 2)
	owner_id_int, _ := strconv.Atoi(data[0])
	user_id_int, _ := strconv.Atoi(data[1])
	owner_id := int64(owner_id_int)
	user_id := int64(user_id_int)
	if c.Sender().ID != user_id && c.Sender().ID != owner_id {
		c.Respond(&tb.CallbackResponse{Text: "This action is not intended for you.", ShowAlert: true})
		return nil
	}
	if c.Sender().ID == user_id {
		c.Edit(fmt.Sprintf("<a href='tg://user?id=%d'>%s</a> has declined the fed transfer.", c.Sender().ID, c.Sender().FirstName))
	} else if c.Sender().ID == owner_id {
		c.Edit(fmt.Sprintf("<a href='tg://user?id=%d'>%s</a> has cancelled the fed transfer.", c.Sender().ID, c.Sender().FirstName))
	}
	return nil
}

func Confirm_Transfer_Fed_cb(c tb.Context) error {
	data := strings.SplitN(c.Callback().Data, "|", 2)
	owner_id_int, _ := strconv.Atoi(data[0])
	user_id_int, _ := strconv.Atoi(data[1])
	owner_id := int64(owner_id_int)
	user_id := int64(user_id_int)
	if c.Sender().ID != owner_id {
		c.Respond(&tb.CallbackResponse{Text: "This action is not intended for you.", ShowAlert: true})
		return nil
	}
	user, _ := c.Bot().ChatByID(user_id)
	_, fed_id, fedname := db.Get_fed_by_owner(owner_id)
	c.Edit(fmt.Sprintf("Congratulations! Federation %s (<code>%s</code>) has successfully been transferred from <a href='tg://user?id=%d'>%s</a> to <a href='tg://user?id=%d'>%s</a>.", fedname, fed_id, c.Sender().ID, c.Sender().FirstName, user.ID, user.FirstName))
	db.Transfer_fed(fed_id, user_id)
	db.User_leave_fed(fed_id, user_id)
	c.Send(fmt.Sprintf(`<b>Fed Transfer</b>
	<b>Fed:</b> %s
	<b>New Fed Owner:</b> <a href='tg://user?id=%d'>%s</a> - <code>%d</code>
	<b>Old Fed Owner:</b> <a href='tg://user?id=%d'>%s</a> - <code>%d</code>
	<a href='tg://user?id=%d'>%s</a> is now the fed owner. They can promote/demote admins as they like.`, fedname, user.ID, user.FirstName, user.ID, c.Sender().ID, c.Sender().FirstName, c.Sender().ID, user.ID, user.FirstName))
	return nil
}

func Deny_Transfer_Fed_cb(c tb.Context) error {
	data := strings.SplitN(c.Callback().Data, "|", 2)
	owner_id_int, _ := strconv.Atoi(data[0])
	owner_id := int64(owner_id_int)
	if c.Sender().ID != owner_id {
		c.Respond(&tb.CallbackResponse{Text: "This action is not intended for you.", ShowAlert: true})
		return nil
	}
	c.Edit(fmt.Sprintf("Fed transfer has been cancelled by <a href='tg://user?id=%d'>%s</a>.", c.Sender().ID, c.Sender().FirstName))
	return nil
}
