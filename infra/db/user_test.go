package db

import (
	"testing"
)

func Test_saveUser(t *testing.T) {
	// Do not run pararell
	r, assert, _ := setupRepo(t, common)
	id := mustNewUUIDV4(t)

	user := &User{
		ID:         id,
		State:      1,
		IcalSecret: "foo",
		Token: Token{
			UserID: id,
			Oauth2Token: &Oauth2Token{
				AccessToken: "hoge",
			},
		},
		Provider: Provider{
			Issuer:  "bar",
			Subject: id.String(),
		},
	}

	t.Run("save user", func(t *testing.T) {
		_, err := saveUser(r.db, user)
		assert.NoError(err)
		u, err := getUser(r.db, id)
		assert.NoError(err)
		assert.Equal(user.Provider.Issuer, u.Provider.Issuer)
	})

	t.Run("Update only state. Without deleting anything.", func(t *testing.T) {
		_, err := saveUser(r.db, &User{
			ID:    user.ID,
			State: 2,
		})
		assert.NoError(err)

		u, err := getUser(r.db.Preload("Token"), id)
		assert.NoError(err)
		// token
		assert.Equal(user.Token.AccessToken, u.Token.AccessToken)
		// provider
		assert.Equal(user.Provider.Issuer, u.Provider.Issuer)
		// icalSecret
		assert.Equal(user.IcalSecret, u.IcalSecret)
	})

	t.Run("Update token", func(t *testing.T) {
		_, err := saveUser(r.db, &User{
			ID:    user.ID,
			State: 2,
			Token: Token{
				Oauth2Token: &Oauth2Token{
					AccessToken: "hoge2",
				},
			},
		})
		assert.NoError(err)
		token, err := getToken(r.db, id)
		assert.NoError(err)
		assert.Equal("hoge2", token.AccessToken)
	})

	t.Run("Update privilege", func(t *testing.T) {
		u, err := getUser(r.db, user.ID)
		assert.NoError(err)
		assert.False(u.Privilege)

		_, err = saveUser(r.db, &User{
			ID:        user.ID,
			Privilege: true,
		})
		assert.NoError(err)

		u, err = getUser(r.db, user.ID)
		assert.NoError(err)
		assert.True(u.Privilege)
	})
}

func Test_getUser(t *testing.T) {
	r, assert, _, user := setupRepoWithUser(t, common)

	t.Run("get user", func(t *testing.T) {
		_, err := getUser(r.db, user.ID)
		assert.NoError(err)
	})
}
