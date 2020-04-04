package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	traQutils "github.com/traPtitech/traQ/utils"
)

func TestGormRepository_CreateOrGetTag(t *testing.T) {
	t.Parallel()
	repo, _, _ := setupGormRepo(t, common)

	tagName := traQutils.RandAlphabetAndNumberString(10)
	tag, err := repo.CreateOrGetTag(tagName)
	assert.NoError(t, err)

	t.Run("Get Tag", func(t *testing.T) {
		if getTag, err := repo.CreateOrGetTag(tagName); assert.NoError(t, err) {
			assert.Equal(t, tag.ID, getTag.ID)
		}
	})

}
