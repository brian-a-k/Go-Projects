package main

import (
	"net/http"
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	log "github.com/llimllib/loglevel"
	"strings"
	"os"
)

var maxDepth int = 2

type Link struct {
	url   string
	text  string
	depth int
}

func (self Link) String() string {
	spacer := strings.Repeat("\t", self.depth)
	return fmt.Sprintf("%s%s (%d) - %s", spacer, self.text, self.depth, self.url)
}

func (self Link) Valid() bool {
	if self.depth >= maxDepth {
		return false
	}

	if len(self.text) == 0 {
		return false
	}

	if len(self.url) == 0 || strings.Contains(strings.ToLower(self.url), "javascript") {
		return false
	}
	return true
}

type HTTpError struct {
	original string
}

func (self HTTpError) Error() string {
	return self.original
}

func LinkReader(resp *http.Response, depth int) []Link {
	page := html.NewTokenizer(resp.Body)
	links := []Link{}

	var start *html.Token
	var text string

	for {
		_ = page.Next()
		token := page.Token()
		if token.Type == html.ErrorToken {
			break
		}

		if start != nil && token.Type == html.TextToken {
			text = fmt.Sprintf("%s%s", text, token.Data)
		}

		if token.DataAtom == atom.A {
			switch token.Type {
			case html.StartTagToken:
				if len(token.Attr) > 0 {
					start = &token
				}
			case html.EndTagToken:
				if start == nil {
					log.Warn("Link end found without start: %s", text)
				}
				link := NewLink(*start, text, depth)
				if link.Valid() {
					links = append(links, link)
					log.Debugf("Link found: %v", link)
				}

				start = nil
				text = ""
			}
		}
	}
	log.Debug(links)
	return links
}

func NewLink(tag html.Token, text string, depth int) Link {
	link := Link{text: strings.TrimSpace(text), depth: depth}
	for i := range tag.Attr {
		if tag.Attr[i].Key == "href" {
			link.url = strings.TrimSpace(tag.Attr[i].Val)
		}
	}
	return link
}

func downLoader(url string) (resp *http.Response, err error) {
	log.Debugf("Downloading %s", url)
	resp, err = http.Get(url)
	if err != nil {
		log.Debugf("Error %s", err)
		return
	}

	if resp.StatusCode > 299 {
		err = HTTpError{fmt.Sprintf("Error (%d): %s", resp.StatusCode, url)}
		log.Debug(err)
		return
	}
	return
}

func recurDownloader(url string, depth int){
	page, err := downLoader(url)

	if err != nil {
		log.Error(err)
		return
	}

	links := LinkReader(page, depth)

	for _, link := range links {
		fmt.Println(link)
		if depth + 1 < maxDepth {
			recurDownloader(link.url, depth + 1)
		}
	}
}


func main()  {
	log.SetPriorityString("info")
	log.SetPrefix("crawler")

	log.Debug(os.Args)

	if len(os.Args) > 2 {
		log.Fatalln("Missing url")
	}
	recurDownloader(os.Args[1], 0)
}
