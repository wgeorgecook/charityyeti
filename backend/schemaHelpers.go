package main

import "github.com/dghubble/go-twitter/twitter"

func newTwitterUser(u *twitter.User) TwitterUser {
	return TwitterUser{
		ID:           u.ID,
		IdString:     u.IDStr,
		Name:         u.Name,
		ScreenName:   u.ScreenName,
		Email:        u.Email,
		Protected:    u.Protected,
		DoNotContact: false,
	}
}
