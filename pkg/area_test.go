package pkg

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAllocateAreas(t *testing.T) {
	areaInfo := AreaInfo{
		Region: "TestArea",
		Areas: []Area{
			{Name: "Area100", Size: 100},
			{Name: "Area75", Size: 75},
			{Name: "Area50", Size: 50},
			{Name: "Area25", Size: 25},
			{Name: "Area20", Size: 20},
			{Name: "Area15", Size: 15},
			{Name: "Area10", Size: 10},
			{Name: "Area5", Size: 5},
		},
	}

	result := CountLinesResult{
		LinesByAuthor: map[string]int{
			"userA": 300,
			"userB": 200,
			"userC": 100,
		},
	}

	testCases := []struct {
		areaInfo    *AreaInfo
		result      *CountLinesResult
		areaAuthors []*AreaAuthor
	}{
		{
			areaInfo: &areaInfo,
			result:   &result,
			areaAuthors: []*AreaAuthor{
				{
					Area:       areaInfo.Areas[0],
					AreaRatio:  0.333,
					Author:     "userA",
					AuthorRank: 1,
				},
				{
					Area:       areaInfo.Areas[1],
					AreaRatio:  0.25,
					Author:     "userB",
					AuthorRank: 2,
				},
				{
					Area:       areaInfo.Areas[2],
					AreaRatio:  0.167,
					Author:     "userA",
					AuthorRank: 1,
				},
				{
					Area:       areaInfo.Areas[3],
					AreaRatio:  0.083,
					Author:     "userB",
					AuthorRank: 2,
				},
				{
					Area:       areaInfo.Areas[4],
					AreaRatio:  0.067,
					Author:     "userC",
					AuthorRank: 3,
				},
				{
					Area:       areaInfo.Areas[5],
					AreaRatio:  0.05,
					Author:     "userC",
					AuthorRank: 3,
				},
				{
					Area:       areaInfo.Areas[6],
					AreaRatio:  0.033,
					Author:     "userC",
					AuthorRank: 3,
				},
				{
					Area:       areaInfo.Areas[7],
					AreaRatio:  0.017,
					Author:     "userC",
					AuthorRank: 3,
				},
			},
		},
	}

	for index, testCase := range testCases {
		t.Run(fmt.Sprintf("case_%v", index), func(t *testing.T) {
			areaAuthors, err := AllocateAreas(testCase.areaInfo, testCase.result)
			assert.NoError(t, err)
			assert.Equal(t, len(testCase.areaInfo.Areas), len(areaAuthors))
			assert.Equal(t, testCase.areaAuthors, areaAuthors)

			totalRatio := float64(0)
			for _, areaAuthor := range areaAuthors {
				totalRatio += areaAuthor.AreaRatio
			}
			assert.Equal(t, float64(1), totalRatio)
		})
	}
}

func TestGetAreaInfo(t *testing.T) {
	areaInfo, err := GetAreaInfo("JP")
	assert.NoError(t, err)

	assert.Equal(t, "JP", areaInfo.Region)
	assert.Equal(t, 47, len(areaInfo.Areas))
	assert.Equal(t, Area{
		Name: "Hokkaido",
		Size: float64(83424),
	}, areaInfo.Areas[0])
}
