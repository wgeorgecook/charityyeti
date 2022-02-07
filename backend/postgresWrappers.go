package main

import (
	"context"
)

// checkDoNotContact takes in a user ID and retreives that user from the database
// if that user is marked DoNotContact, returns true. Else, false.
func checkDoNotContact(ctx context.Context, userID string) bool {
	log.Infof("Checking if user %v is do not contact", userID)
	// find the record in the database
	user, err := getDonor(ctx, userID)
	if err != nil {
		// if we get an error here, we should assume that we do not
		// want to contact this uer
		log.Errorf("could not check user: %v", err)
		return false
	}
	log.Debugf("found user: %+v", user)
	return user.DoNotContact
}

// addDoNotContact takes a user ID that wishes to be marked Do Not Contact
// and updates their user record accordingly
func addDoNotContact(ctx context.Context, userID string) error {
	log.Infof("Marking %v do not contact", userID)

	// get user data from database
	user, err := getDonor(ctx, userID)
	if err != nil {
		log.Errorf("could not get user to mark as do not contact: %v", err)
		return err
	}

	if user.DoNotContact {
		// no need to re-apply
		return nil
	}

	// otherwise
	user.DoNotContact = true
	if err := updateDonor(ctx, user); err != nil {
		// bork
		log.Errorf("could not update user as do not contact: %v", err)
		return err
	}
	// we done
	return nil
}

// removeDoNotContact takes a user ID that wishes to be un-marked as Do Not Contact
// and updates their user record accordingly
func removeDoNotContact(ctx context.Context, userID string) error {
	log.Infof("Unmarking %v as do not contact", userID)

	// get user data from database
	user, err := getDonor(ctx, userID)
	if err != nil {
		log.Errorf("could not get user to remove do not contact: %v", err)
		return err
	}

	if !user.DoNotContact {
		// no need to re-apply
		return nil
	}

	// otherwise
	user.DoNotContact = false
	if err := updateDonor(ctx, user); err != nil {
		// bork
		log.Errorf("could not update user as do not contact: %v", err)
		return err
	}
	// we done
	return nil
}
