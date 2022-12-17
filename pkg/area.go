package pkg

import (
	"errors"
	"fmt"
	"log"
	"math"
	"sort"
)

type AreaAuthor struct {
	Area      Area
	AreaRatio float64
	Author    string
}

func Kunitori(areaInfo *AreaInfo, result *CountLinesResult) ([]*AreaAuthor, error) {
	type rank struct {
		author     string
		lines      int
		linesRatio float64
		areaRatio  float64
	}

	log.Printf("start GetAreaAuthors: areaInfo=%+v, result=%+v", areaInfo, result)

	if len(result.LinesByAuthor) == 0 {
		return nil, errors.New("CountLinesResult is empty")
	}

	totalAuthors, totalLines := 0, 0
	for _, lines := range result.LinesByAuthor {
		totalLines += lines
		totalAuthors++
	}

	log.Printf("count: totalAuthors=%v, totalLines=%v", totalAuthors, totalLines)

	ranks := make([]rank, 0)
	for author, lines := range result.LinesByAuthor {
		ranks = append(ranks, rank{
			author:     author,
			lines:      lines,
			linesRatio: float64(lines) / float64(totalLines),
			areaRatio:  0,
		})
	}

	sort.SliceStable(ranks, func(i, j int) bool {
		if ranks[i].lines == ranks[j].lines {
			return ranks[i].author < ranks[j].author
		} else {
			return ranks[i].lines > ranks[j].lines
		}
	})

	log.Printf("ranks: count=%+v", len(ranks))

	totalAreaSize := float64(0)
	for _, area := range areaInfo.Areas {
		totalAreaSize += area.Size
	}

	log.Printf("count: areaCount=%v totalAreaSize=%v", len(areaInfo.Areas), totalAreaSize)

	log.Printf("start kunitori!")

	areaAuthors := make([]*AreaAuthor, 0)
	fraction := float64(1)
	for _, area := range areaInfo.Areas {
		areaRatio := area.Size / totalAreaSize
		author := ""
		for index, rank := range ranks {
			if rank.linesRatio >= rank.areaRatio+areaRatio {
				log.Printf(
					"allocate: area=%v, areaRatio=%.2f => author=%v, linesRatio=%.2f, authorAreaRatio=%.2f => %.2f",
					area.Name,
					areaRatio,
					rank.author,
					rank.linesRatio,
					rank.areaRatio,
					rank.areaRatio+areaRatio,
				)

				author = rank.author
				ranks[index].areaRatio += areaRatio
				break
			}
		}

		if author == "" {
			log.Printf("skip: area=%v, areaRatio=%v", area.Name, areaInfo)
		}

		roundedAreaRatio := math.Round(areaRatio*1000) / 1000
		areaAuthors = append(areaAuthors, &AreaAuthor{
			Area:      area,
			Author:    author,
			AreaRatio: roundedAreaRatio,
		})
		fraction -= roundedAreaRatio
	}

	roundedFraction := math.Round(fraction*1000) / 1000
	areaAuthors[len(areaAuthors)-1].AreaRatio += roundedFraction

	log.Printf("complete kunitori: areaAuthors=%v, roundedFraction=%v", len(areaAuthors), roundedFraction)

	return areaAuthors, nil
}

type Area struct {
	Name string
	Size float64
}

type AreaInfo struct {
	Region string
	Areas  []Area
}

func GetAreaInfo(region string) (*AreaInfo, error) {
	switch region {
	case "JP":
		return &AreaInfo{
			Region: region,
			Areas: []Area{
				{Name: "Hokkaido", Size: 83424},
				{Name: "Aomori", Size: 9646},
				{Name: "Iwate", Size: 15275},
				{Name: "Miyagi", Size: 7282},
				{Name: "Akita", Size: 11638},
				{Name: "Yamagata", Size: 9323},
				{Name: "Fukushima", Size: 13784},
				{Name: "Ibaraki", Size: 6097},
				{Name: "Tochigi", Size: 6408},
				{Name: "Gunma", Size: 6362},
				{Name: "Saitama", Size: 3798},
				{Name: "Chiba", Size: 5158},
				{Name: "Tokyo", Size: 2191},
				{Name: "Kanagawa", Size: 2416},
				{Name: "Niigata", Size: 12584},
				{Name: "Toyama", Size: 4248},
				{Name: "Ishikawa", Size: 4186},
				{Name: "Fukui", Size: 4190},
				{Name: "Yamanashi", Size: 4465},
				{Name: "Nagano", Size: 13562},
				{Name: "Gifu", Size: 10621},
				{Name: "Shizuoka", Size: 7777},
				{Name: "Aichi", Size: 5172},
				{Name: "Mie", Size: 5774},
				{Name: "Shiga", Size: 4017},
				{Name: "Kyoto", Size: 4612},
				{Name: "Osaka", Size: 1905},
				{Name: "Hyogo", Size: 8401},
				{Name: "Nara", Size: 3691},
				{Name: "Wakayama", Size: 4725},
				{Name: "Tottori", Size: 3507},
				{Name: "Shimane", Size: 6708},
				{Name: "Okayama", Size: 7115},
				{Name: "Hiroshima", Size: 8479},
				{Name: "Yamaguchi", Size: 6112},
				{Name: "Tokushima", Size: 4147},
				{Name: "Kagawa", Size: 1877},
				{Name: "Ehime", Size: 5676},
				{Name: "Kouchi", Size: 7104},
				{Name: "Fukuoka", Size: 4986},
				{Name: "Saga", Size: 2441},
				{Name: "Nagasaki", Size: 4132},
				{Name: "Kumamoto", Size: 7409},
				{Name: "Oita", Size: 6341},
				{Name: "Miyazaki", Size: 7735},
				{Name: "Kagoshima", Size: 9187},
				{Name: "Okinawa", Size: 2281},
			},
		}, nil
	}

	return nil, fmt.Errorf("not found: region=%v", region)
}
