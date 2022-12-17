package pkg

import (
	"bytes"
	_ "embed"
	"html/template"
)

//go:embed chart.html
var chartHtml string

func RenderChartHtml(generateResult *GenerateResult) (string, error) {
	chartHtmlTemplate, err := template.New("chart").Parse(chartHtml)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBufferString("")
	err = chartHtmlTemplate.Execute(buf, generateResult)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
