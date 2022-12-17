package pkg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAreaInfo(t *testing.T) {
	areaInfo, err := GetAreaInfo("JP")
	assert.NoError(t, err)

	assert.Equal(t, "JP", areaInfo.Region)
	assert.Equal(t, 47, len(areaInfo.Areas))
	assert.Equal(t, Area{
		Name: "北海道",
		Size: 83424,
	}, areaInfo.Areas[0])
}
