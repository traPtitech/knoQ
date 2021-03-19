package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func Test_saveUser(t *testing.T) {
	// Do not run pararell
	r, _, _ := setupRepo(t, common)
	id := mustNewUUIDV4(t)

	user := &User{
		ID:         id,
		State:      1,
		IcalSecret: "foo",
		Token: Token{
			UserID: id,
			Token: &oauth2.Token{
				AccessToken: "hoge",
			},
		},
		Provider: Provider{
			UserID:  id,
			Issuer:  "bar",
			Subject: id.String(),
		},
	}

	t.Run("save user", func(t *testing.T) {
		_, err := saveUser(r.db, user)
		assert.NoError(t, err)
		u, err := getUser(r.db.Preload("Provider"), id)
		assert.NoError(t, err)
		assert.Equal(t, user.Provider.Issuer, u.Provider.Issuer)
	})

	t.Run("Update only state. Without deleting anything.", func(t *testing.T) {
		_, err := saveUser(r.db, &User{
			ID:    user.ID,
			State: 2,
		})
		assert.NoError(t, err)

		u, err := getUser(r.db.Preload("Token").Preload("Provider"), id)
		assert.NoError(t, err)
		// token
		assert.Equal(t, user.Token.AccessToken, u.Token.AccessToken)
		// provider
		assert.Equal(t, user.Provider.Issuer, u.Provider.Issuer)
		// icalSecret
		assert.Equal(t, user.IcalSecret, u.IcalSecret)
	})

	t.Run("Update token", func(t *testing.T) {
		_, err := saveUser(r.db, &User{
			ID:    user.ID,
			State: 2,
			Token: Token{
				Token: &oauth2.Token{
					AccessToken: "hoge2",
				},
			},
		})
		assert.NoError(t, err)
		token, err := getToken(r.db, id)
		assert.NoError(t, err)
		assert.Equal(t, "hoge2", token.AccessToken)
	})
}

func Test_getUser(t *testing.T) {
	r, _, _, user := setupRepoWithUser(t, common)

	t.Run("get user", func(t *testing.T) {
		_, err := getUser(r.db, user.ID)
		assert.NoError(t, err)
	})
}
