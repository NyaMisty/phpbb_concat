package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

func getPageContent(urlStr string, pageNumber int) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	// Modify start parameter in query string
	query := parsedURL.Query()
	query.Set("start", strconv.Itoa(pageNumber*50))
	parsedURL.RawQuery = query.Encode()

	response, err := http.Get(parsedURL.String())
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func extractTotalPages(pageContent string) (int, error) {
	re := regexp.MustCompile(`<div class="page-number">Page <strong>\d+</strong> of <strong>(\d+)</strong></div>`)
	match := re.FindStringSubmatch(pageContent)
	if match == nil || len(match) < 2 {
		return 0, fmt.Errorf("total pages not found")
	}

	totalPages, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, err
	}

	return totalPages, nil
}

func splitPage(pageContent string) (string, string, string) {
	// 从页面内容中提取需要的部分
	startIndex := strings.Index(pageContent, `<div id="page-body" class="page-body">`)
	if startIndex == -1 {
		panic("djefhrvb")
	}
	endIndex := strings.Index(pageContent[startIndex:], `<div id="page-footer" class="page-footer">`)
	if endIndex == -1 {
		panic("cnklejebw")
	}
	endIndex += startIndex

	return pageContent[:startIndex], pageContent[startIndex:endIndex], pageContent[endIndex:]
}

func fetchPageContent(urlStr string, w io.Writer) error {
	write := func(s string) {
		w.Write([]byte(s))
	}

	firstPageContent, err := getPageContent(urlStr, 0)
	if err != nil {
		return err
	}
	log.Printf("got first page content for %s", urlStr)

	pageHdr, firstPageBody, pageEnd := splitPage(firstPageContent)

	totalPages, err := extractTotalPages(firstPageContent)
	if err != nil {
		return err
	}
	log.Printf("thread %s has %d pages", urlStr, totalPages)

	pages := make([]string, totalPages)
	pages[0] = firstPageBody

	var egFin bool

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		write(pageHdr)
		for i, _ := range pages {
			for pages[i] == "" && !egFin {
				time.Sleep(1 * time.Second)
			}
			pageBody := pages[i]
			if pageBody == "" {
				break
			}
			log.Printf("thread %s page %d written", urlStr, i)
			write(pageBody)
		}
		write(pageEnd)
	}()

	var eg errgroup.Group

	eg.SetLimit(10)
	// 循环请求页面，拼接内容
	for pageNumber := 1; pageNumber < totalPages; pageNumber++ { // 假设总共有6页
		pageNumber := pageNumber // Capture the loop variable
		eg.Go(func() error {
			pageContent, err := getPageContent(urlStr, pageNumber)
			if err != nil {
				return err
			}

			_, pageBodyContent, _ := splitPage(pageContent)
			pages[pageNumber] = pageBodyContent

			log.Printf("thread %s page %d finished", urlStr, pageNumber)
			return nil
		})
	}

	go func() {
		err = eg.Wait()
		egFin = true
		log.Printf("thread %s fully finished", urlStr)
	}()

	wg.Wait()

	if err != nil {
		return err
	}

	return nil
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// urlStr := "https://www.bogleheads.org/forum/viewtopic.php?t=288192&view=print"
	urlStr := r.URL.Query().Get("url")
	if urlStr == "" {
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	// Fetch page content
	err := fetchPageContent(urlStr, w)
	if err != nil {
		http.Error(w, "Failed to fetch page content", http.StatusInternalServerError)
		return
	}
}

func main() {
	if true {
		port := 3322
		http.HandleFunc("/", handleRequest)
		fmt.Printf("Server listening on port %v...", port)
		http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	} else {
		err := fetchPageContent("https://www.bogleheads.org/forum/viewtopic.php?t=288192&view=print", io.Discard)
		_ = err
	}
}
