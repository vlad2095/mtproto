package mtproto

import (
	"fmt"
	"log"
)

// Contact
type Contact struct {
	ClientID  int64
	Firstname string
	Lastname  string
	Phone     string
	UserID    int32
	Mutual    bool
}

func (c *Contact) GetInputContact() TL {
	return TL_inputPhoneContact{
		Client_id:  c.ClientID,
		First_name: c.Firstname,
		Last_name:  c.Lastname,
		Phone:      c.Phone,
	}
}

func NewContact(in TL) (contact *Contact) {
	contact = new(Contact)
	switch c := in.(type) {
	case TL_contact:
		contact.UserID = c.User_id
		contact.Mutual = toBool(c.Mutual)
	case TL_importedContact:
		contact.UserID = c.User_id
		contact.ClientID = c.Client_id
	case TL_inputPhoneContact:
		contact.ClientID = c.Client_id
		contact.Firstname = c.First_name
		contact.Lastname = c.Last_name
		contact.Phone = c.Phone
	default:
		log.Println("GetContact::Error::Invalid Type")
		return nil
	}
	return
}

func (m *MTProto) Contacts_ResolveUserName(name string) ([]Channel, []Chat, []User, error) {
	resp := make(chan TL, 1)
	m.queueSend <- packetToSend{TL_contacts_resolveUsername{
		name},
		resp,
	}
	x := <-resp

	peer, ok := x.(TL_contacts_resolvedPeer)
	if !ok {
		log.Println(fmt.Sprintf("RPC: %#v", x))
		return []Channel{}, []Chat{}, []User{}, fmt.Errorf("RPC: %#v", x)
	}

	TChats := make([]Chat, 0, len(peer.Chats))
	TChannel := make([]Channel, 0, len(peer.Chats))
	TUsers := make([]User, 0, len(peer.Users))

	for _, v := range peer.Chats {
		switch v.(type) {
		case TL_chatEmpty, TL_chat, TL_chatFull, TL_chatForbidden:
			TChats = append(
				TChats,
				*NewChat(v),
			)
		case TL_channel, TL_channelFull, TL_channelForbidden:
			TChannel = append(
				TChannel,
				*NewChannel(v),
			)
		}
	}

	for _, v := range peer.Users {
		switch u := v.(type) {
		case TL_user, TL_userEmpty:
			TUsers = append(TUsers, *NewUser(u))
		case TL_userProfilePhoto:
			TUsers[len(TUsers)-1].Photo = NewUserProfilePhoto(u)
		case TL_userStatusRecently, TL_userStatusOffline, TL_userStatusOnline, TL_userStatusLastWeek, TL_userStatusLastMonth:
			TUsers[len(TUsers)-1].Status = NewUserStatus(u)
		}
	}

	return TChannel, TChats, TUsers, nil
}

func (m *MTProto) Contacts_GetContacts(hash int32) ([]Contact, []User, error) {
	resp := make(chan TL, 1)
	m.queueSend <- packetToSend{TL_contacts_getContacts{
		hash},
		resp,
	}
	x := <-resp
	list, ok := x.(TL_contacts_contacts)
	if !ok {
		log.Println(fmt.Sprintf("RPC: %#v", x))
		return []Contact{}, []User{}, fmt.Errorf("RPC: %#v", x)
	}
	TContacts := make([]Contact, 0, len(list.Contacts))
	TUsers := make([]User, 0, len(list.Users))
	for _, v := range list.Contacts {
		TContacts = append(
			TContacts,
			*NewContact(v),
		)
	}
	for _, v := range list.Users {
		switch u := v.(type) {
		case TL_user, TL_userEmpty:
			TUsers = append(TUsers, *NewUser(u))
		case TL_userProfilePhoto:
			TUsers[len(TUsers)-1].Photo = NewUserProfilePhoto(u)
		case TL_userStatusRecently, TL_userStatusOffline, TL_userStatusOnline, TL_userStatusLastWeek, TL_userStatusLastMonth:
			TUsers[len(TUsers)-1].Status = NewUserStatus(u)
		}
	}
	return TContacts, TUsers, nil
}

func (m *MTProto) Contacts_ImportContacts(contacts []TL) {
	resp := make(chan TL, 1)
	m.queueSend <- packetToSend{
		TL_contacts_importContacts{
			contacts,
		},
		resp,
	}
	x := <-resp
	switch r := x.(type) {
	case TL_contacts_importedContacts:
		//TODO:: must do something with response
		log.Println(r)
	default:
		log.Println(fmt.Sprintf("RPC: %#v", x))
		return

	}

}
