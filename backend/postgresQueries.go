package main

import (
	"context"
)

// createDonation inserts this donation into our database
func createDonation(ctx context.Context, d *Donation) error {
	log.Infof("inserting donation %v", d.ID)
	if _, err := pgClient.NewInsert().Model(d).Exec(ctx); err != nil {
		log.Errorf("could not insert donation: %v", err)
		return err
	}
	log.Info("donation inserted")
	return nil
}

// getDonation retreives the donation with the provided ID
func getDonation(ctx context.Context, id string) (*Donation, error) {
	log.Infof("querying for donation %v", id)

	// start our pg query
	d := new(Donation)
	err := pgClient.NewSelect().Model(&d).Where("id = ?", id).Scan(ctx)
	if err != nil {
		log.Errorf("could not get donation: %v", err)
		return nil, err
	}
	log.Info("returning donation")
	return d, nil
}

// updateDonation takes an entire donation and writes it completely
// to the database, replacing the donation of the same id
func updateDonation(ctx context.Context, d *Donation) error {
	log.Infof("updating donation %v", d.ID)
	_, err := pgClient.NewUpdate().Model(d).Where("id = ?", d.ID).Returning("*").Exec(ctx)
	if err != nil {
		log.Errorf("could not update donation: %v", err)
		return err
	}
	log.Infof("complete update donation %v", d.ID)
	return nil
}

// createDonor inserts this donating user into our database
func createDonor(ctx context.Context, d *Donor) error {
	log.Infof("inserting donor %v", d.ID)
	if _, err := pgClient.NewInsert().Model(d).Exec(ctx); err != nil {
		log.Errorf("could not insert donor: %v", err)
		return err
	}
	log.Infof("donor inserted")
	return nil
}

// getDonor retreives the user with the provided ID
func getDonor(ctx context.Context, id string) (*Donor, error) {
	log.Infof("querying for donor %v", id)
	// start our pg query
	u := new(Donor)
	err := pgClient.NewSelect().Model(&u).Where("id = ?", id).Scan(ctx)
	if err != nil {
		log.Errorf("could not query for donor: %v", err)
		return nil, err
	}
	log.Info("returning donor")
	return u, nil
}

// updateDonor takes an entire user and writes it completely
// to the database, replacing the user of the same id
func updateDonor(ctx context.Context, d *Donor) error {
	log.Infof("updating user %v", d.ID)
	_, err := pgClient.NewUpdate().Model(d).Where("id = ?", d.ID).Returning("*").Exec(ctx)
	if err != nil {
		log.Errorf("could not update donor: %v", err)
		return err
	}
	log.Infof("complete update donor %v", d.ID)
	return nil
}

// createHonorary inserts this honored user into our database
func createHonorary(ctx context.Context, h *Honorary) error {
	log.Infof("inserting honorary %v", h.ID)
	if _, err := pgClient.NewInsert().Model(h).Exec(ctx); err != nil {
		log.Errorf("could not insert honorary: %v", err)
		return err
	}
	log.Infof("honorary inserted")
	return nil
}

// getHonorary retreives the honorary with the provided ID
func getHonorary(ctx context.Context, id string) (*Honorary, error) {
	log.Infof("querying for honorary %v", id)
	// start our pg query
	u := new(Honorary)
	err := pgClient.NewSelect().Model(&u).Where("id = ?", id).Scan(ctx)
	if err != nil {
		log.Errorf("could not query for donor: %v", err)
		return nil, err
	}
	log.Info("returning donor")
	return u, nil
}
