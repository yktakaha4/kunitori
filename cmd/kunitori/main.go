package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dlclark/regexp2"
	"github.com/yktakaha4/kunitori/pkg"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ", ")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	if os.Getenv("DEBUG") == "" {
		log.SetOutput(io.Discard)
	}

	defaultHelpMessage := `Kunitori

[subcommands]
generate	...	generate Kunitori chart
`

	if len(os.Args) < 2 {
		fmt.Print(defaultHelpMessage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "generate":
		generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
		generateOut := generateCmd.String("out", ".", "out directory path")
		generateJson := generateCmd.Bool("json", false, "export as json format")
		generateUrl := generateCmd.String("url", "", "repository url")
		generatePath := generateCmd.String("path", "", "repository path")
		generateRegion := generateCmd.String("region", "JP", "chart region")
		generateSince := generateCmd.String(
			"since",
			"",
			fmt.Sprintf("filter commit since date (format: %v)", time.RFC3339),
		)
		generateUntil := generateCmd.String(
			"until",
			"",
			fmt.Sprintf("filter commit until date (format: %v)", time.RFC3339),
		)
		generateInterval := generateCmd.Duration(
			"interval",
			time.Hour*24*30,
			"commit pick interval",
		)
		generateLimit := generateCmd.Int(
			"limit",
			12,
			"commit pick limit",
		)

		var filters arrayFlags
		generateCmd.Var(
			&filters,
			"filters",
			"target file filter regex (multiple specified)",
		)

		var authors arrayFlags
		generateCmd.Var(
			&authors,
			"authors",
			"target file author regex (multiple specified, format: author=regex)",
		)

		err := generateCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		since, until := time.UnixMilli(0).UTC(), time.Now().UTC()
		if *generateSince != "" {
			since, err = time.Parse(time.RFC3339, *generateSince)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		if *generateUntil != "" {
			until, err = time.Parse(time.RFC3339, *generateUntil)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		filterRegexes := make([]*regexp2.Regexp, 0)
		for _, filter := range filters {
			regex, err := regexp2.Compile(filter, 0)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			filterRegexes = append(filterRegexes, regex)
		}

		if len(filterRegexes) == 0 {
			filterRegexes = append(filterRegexes, regexp2.MustCompile(".+", 0))
		}

		authorRegexes := make([]pkg.AuthorRegex, 0)
		for _, author := range authors {
			parts := strings.Split(author, "=")
			if len(parts) != 2 {
				fmt.Println(fmt.Sprintf("invalid format: %v", authors))
				os.Exit(1)
			}

			regex, err := regexp2.Compile(parts[1], 0)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			authorRegexes = append(authorRegexes, pkg.AuthorRegex{
				Condition: regex,
				Author:    parts[0],
			})
		}

		if _, err := os.Stat(*generateOut); os.IsNotExist(err) {
			fmt.Println(err)
			os.Exit(1)
		}

		options := pkg.GenerateOptions{
			RepositoryUrl:  *generateUrl,
			RepositoryPath: *generatePath,
			Region:         *generateRegion,
			SearchCommitsOptions: &pkg.SearchCommitsOptions{
				Since:    since,
				Until:    until,
				Interval: *generateInterval,
				Limit:    *generateLimit,
			},
			CountLinesOption: &pkg.CountLinesOption{
				Filters:       filterRegexes,
				AuthorRegexes: authorRegexes,
			},
		}

		generateResult, err := pkg.Generate(&options)
		if err != nil {
			panic(err)
		}

		var fileName string
		var data []byte

		if *generateJson {
			fileName = path.Join(*generateOut, "generate.json")
			data, err = json.Marshal(generateResult)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			fileName = path.Join(*generateOut, "chart.html")
			html, err := pkg.RenderChartHtml(generateResult)
			if err != nil {
				panic(err)
			}
			data = []byte(html)
		}

		absFileName, err := filepath.Abs(fileName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		f, err := os.Create(absFileName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		_, err = f.Write(data)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("output: %v", absFileName)
		os.Exit(0)
	default:
		fmt.Print(defaultHelpMessage)
		os.Exit(1)
	}
}
