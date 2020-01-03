package repository

import "errors"

func (us *UserSession) Create() error {
	err := DB.Create(&us).Error
	if err != nil {
		dbErrorLog(err)
		return err
	}
	return nil
}

func (us *UserSession) Get() error {
	if us.Token == "" {
		return errors.New("token is nil")
	}

	err := DB.Take(&us).Error
	if err != nil {
		dbErrorLog(err)
		return err
	}
	return nil
}

func (us *UserSession) Update() error {
	if us.Token == "" {
		return errors.New("token is nil")
	}

	err := DB.Save(&us).Error
	if err != nil {
		dbErrorLog(err)
		return err
	}
	return nil

}
