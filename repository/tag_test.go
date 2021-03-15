package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	traQrandom "github.com/traPtitech/traQ/utils/random"
)

func TestGormRepository_CreateOrGetTag(t *testing.T) {
	t.Parallel()
	repo, _, _ := setupGormRepo(t, common)

	tagName := traQrandom.AlphaNumeric(10)
	tag, err := repo.CreateOrGetTag(tagName)
	assert.NoError(t, err)

	t.Run("Get Tag", func(t *testing.T) {
		if getTag, err := repo.CreateOrGetTag(tagName); assert.NoError(t, err) {
			assert.Equal(t, tag.ID, getTag.ID)
		}
	})

}
