package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: torrentLookup [search string]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func GetContents_1377x(url string) int {
	res, err := http.Get(url)
	if err != nil {
		panic("ERROR")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		fmt.Println("Could not reach site")
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		panic("ERROR")
	}
	results := 0
	// gets number of links on page 1
	doc.Find(".coll-1.name").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find("a").Text())
		if len(title) < 1 {
			return
		}
		results = i
	})
	// gets number of pages
	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages := strings.TrimSpace(strings.Split(s.Text(), ">>")[1])
		//fmt.Println()
		pageNum, err := strconv.Atoi(pages)
		if err != nil {
			panic("ERROR")
		}
		results = results * (pageNum - 1)
	})
	return results
}

func GetContents_ThePirateBay(url string) int {
	res, err := http.Get(url)
	if err != nil {
		panic("ERROR")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		fmt.Println("Could not reach site")
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		panic("ERROR")
	}
	results := 0
	doc.Find("a.detLink").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Text())
		if len(title) < 1 {
			return
		}
		results = i
	})

	doc.Find("td").Each(func(i int, s *goquery.Selection) {
		pages, ok := s.Attr("colspan")
		if !ok {
			return
		}
		if pages == "9" {
			pageList := strings.TrimSpace(s.Text())
			maxPage, err := strconv.Atoi(strings.Split(pageList, "    ")[1])
			if err != nil {
				panic(fmt.Errorf("ERROR: %w", err))
			}
			results = results * (maxPage - 1)
		}
	})
	return results
}

func SendResult(name string, results int, resultCh chan Result) {
	resultCh <- Result{Name: name, hits: results}
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Search String Missing!")
		os.Exit(1)
	}
	searchString := args[0]
	resultCh := make(chan Result)
	sites := 2
	// 1377x
	searchUrl := fmt.Sprintf("https://www.1377x.to/search/%s/1/", searchString)
	go SendResult("1377x", GetContents_1377x(searchUrl), resultCh)

	// ThePirateBay
	searchUrl = fmt.Sprintf("https://www1.thepiratebay3.to/s/?q=%s", searchString)
	go SendResult("ThePirateBay", GetContents_ThePirateBay(searchUrl), resultCh)

	for i := 0; i < sites; i++ {
		select {
		case result := <-resultCh:
			fmt.Printf("%s has ~%v hits\n", result.Name, result.hits)
		}
	}
}

type Result struct {
	Name string
	hits int
}
