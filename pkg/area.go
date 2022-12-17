package pkg

import (
	"fmt"
)

type Area struct {
	Name string
	Size int
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
				{Name: "北海道", Size: 83424},
				{Name: "青森", Size: 9646},
				{Name: "岩手", Size: 15275},
				{Name: "宮城", Size: 7282},
				{Name: "秋田", Size: 11638},
				{Name: "山形", Size: 9323},
				{Name: "福島", Size: 13784},
				{Name: "茨城", Size: 6097},
				{Name: "栃木", Size: 6408},
				{Name: "群馬", Size: 6362},
				{Name: "埼玉", Size: 3798},
				{Name: "千葉", Size: 5158},
				{Name: "東京", Size: 2191},
				{Name: "神奈川", Size: 2416},
				{Name: "新潟", Size: 12584},
				{Name: "富山", Size: 4248},
				{Name: "石川", Size: 4186},
				{Name: "福井", Size: 4190},
				{Name: "山梨", Size: 4465},
				{Name: "長野", Size: 13562},
				{Name: "岐阜", Size: 10621},
				{Name: "静岡", Size: 7777},
				{Name: "愛知", Size: 5172},
				{Name: "三重", Size: 5774},
				{Name: "滋賀", Size: 4017},
				{Name: "京都", Size: 4612},
				{Name: "大阪", Size: 1905},
				{Name: "兵庫", Size: 8401},
				{Name: "奈良", Size: 3691},
				{Name: "和歌山", Size: 4725},
				{Name: "鳥取", Size: 3507},
				{Name: "島根", Size: 6708},
				{Name: "岡山", Size: 7115},
				{Name: "広島", Size: 8479},
				{Name: "山口", Size: 6112},
				{Name: "徳島", Size: 4147},
				{Name: "香川", Size: 1877},
				{Name: "愛媛", Size: 5676},
				{Name: "高知", Size: 7104},
				{Name: "福岡", Size: 4986},
				{Name: "佐賀", Size: 2441},
				{Name: "長崎", Size: 4132},
				{Name: "熊本", Size: 7409},
				{Name: "大分", Size: 6341},
				{Name: "宮崎", Size: 7735},
				{Name: "鹿児島", Size: 9187},
				{Name: "沖縄", Size: 2281},
			},
		}, nil
	}

	return nil, fmt.Errorf("not found: region=%v", region)
}
