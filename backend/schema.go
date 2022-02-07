package main

import (
	"time"

	"github.com/uptrace/bun"
)

// TwitterUser is an exension of a go-twitter user so we can
// implement our own helper methods on it.
type TwitterUser struct {
	ID           int64 `bun:",pk"`
	IdString     string
	Name         string
	ScreenName   string
	Email        string
	Protected    bool
	DoNotContact bool
}

// Donor is a twitterUser who invoked Charity Yeti and is
// donating on behalf of an honorary. Donors have a
// one-to-many relationship with both donations and
// beneficiaries.
type Donor struct {
	bun.BaseModel `bun:"table:donor"`
	TwitterUser
}

// Honorary is a twitterUser whose tweet prompted a
// donor to invoke Charity Yeti. Beneficiaries have a
// one-to-many relationship with donors and donations
type Honorary struct {
	bun.BaseModel `bun:"table:honorary"`
	TwitterUser
}

// donation is the data we store after a donor successfully
// makes a monetary donation on behalf of a honorary's tweet.
// Donations have a one-to-one relationship with both donors
// and beneficiaries.
type Donation struct {
	ID                     string    `json:"id" bun:",pk,type:uuid,unique"`
	OriginalTweetID        int64     `json:"originalTweetID,omitempty"`
	InvokerTweetID         int64     `json:"invokerTweetID,omitempty"`
	Donor                  *Donor    `bun:"rel:belongs-to,join:donor_id=id"`
	Honorary               *Honorary `bun:"rel:belongs-to,join:honorary_id=id"`
	DonationValue          float32   `json:"donationValue,omitempty"`
	DonationID             string    `json:"donationID,omitempty"`
	InvokerResponseTweetID int64     `json:"invokerResponseTweetID,omitempty"`
	CreatedAt              time.Time `json:"createdAt"`
}
