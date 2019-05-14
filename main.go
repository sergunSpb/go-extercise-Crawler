package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	Fetch(url string) (body string, urls []string, err error)
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, wg *sync.WaitGroup , visitMap *syncVisitedMap) {
	defer wg.Done()

    if depth <= 0 {
		return
	}
	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	visitMap.Visit(url)
	fmt.Printf("found: %s %q\n", url, body)
	
	for _, u := range urls {
		if visitMap.IsVisited(u) {
			continue
		}
		wg.Add(1)		
		go Crawl(u, depth-1, fetcher,wg,visitMap)
	}
	return
}

func main() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	visitMap := syncVisitedMap{ make(map[string]*bool) , sync.Mutex{} }
	go Crawl("https://golang.org/", 4, fetcher , &wg , &visitMap)
	wg.Wait()
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}


type syncVisitedMap struct{
	innerMap map[string]*bool
	mux sync.Mutex
}

func(f *syncVisitedMap) Visit(v string){
	f.mux.Lock()
	f.innerMap[v] = nil
	f.mux.Unlock()
}

func(f *syncVisitedMap) IsVisited(v string) bool{
	f.mux.Lock()
	defer f.mux.Unlock()
	_, ok := f.innerMap[v]
	return ok
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
