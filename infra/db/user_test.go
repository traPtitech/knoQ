package db

import (
	"testing"
)

func Test_saveUser(t *testing.T) {
	// Do not run pararell
	r, assert, _ := setupRepo(t, common)
	id := mustNewUUIDV4(t)

	user := &User{
		ID:           id,
		State:        1,
		IcalSecret:   "foo",
		AccessToken:  "hoge",
		ProviderName: "bar",
	}

	t.Run("save user", func(_ *testing.T) {
		_, err := saveUser(r.db, user)
		assert.NoError(err)
		u, err := getUser(r.db, id)
		assert.NoError(err)
		assert.Equal(user.ProviderName, u.ProviderName)
	})

	t.Run("Update only state. Without deleting anything.", func(_ *testing.T) {
		_, err := saveUser(r.db, &User{
			ID:    user.ID,
			State: 2,
		})
		assert.NoError(err)

		u, err := getUser(r.db, id)
		assert.NoError(err)
		// token
		assert.Equal(user.AccessToken, u.AccessToken)
		// provider
		assert.Equal(user.ProviderName, u.ProviderName)
		// icalSecret
		assert.Equal(user.IcalSecret, u.IcalSecret)
	})

	t.Run("Update token", func(_ *testing.T) {
		_, err := saveUser(r.db, &User{
			ID:          user.ID,
			State:       2,
			AccessToken: "hoge2",
		})
		assert.NoError(err)
		token, err := getToken(r.db, id)
		assert.NoError(err)
		assert.Equal("hoge2", token.AccessToken)
	})

	t.Run("Update privilege", func(_ *testing.T) {
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

	t.Run("get user", func(_ *testing.T) {
		_, err := getUser(r.db, user.ID)
		assert.NoError(err)
	})
}
