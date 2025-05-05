package google

import (
	"context"

	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

func (repo *Repository) GetUser(token *oauth2.Token) (*idtoken.Payload, error) {
	idToken := token.Extra("id_token").(string)
	return idtoken.Validate(context.TODO(), idToken, "")
}
