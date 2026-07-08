package integration_test

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/html"
)

func baseURL() string {
	if u := os.Getenv("BASE_URL"); u != "" {
		return strings.TrimRight(u, "/")
	}
	return "https://andrewwillette.com"
}

func TestSheetMusicLinks(t *testing.T) {
	client := &http.Client{Timeout: 15 * time.Second}

	pageURL := baseURL() + "/sheet-music"
	resp, err := client.Get(pageURL)
	if err != nil {
		t.Fatalf("GET %s: %v", pageURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s returned %d", pageURL, resp.StatusCode)
	}

	links := extractHrefs(resp)
	if len(links) == 0 {
		t.Fatal("no links found on sheet music page")
	}

	t.Logf("found %d links on %s", len(links), pageURL)

	base := baseURL()
	failed := 0
	for _, link := range links {
		if strings.HasPrefix(link, "/") {
			link = base + link
		}
		linkResp, err := client.Get(link)
		if err != nil {
			t.Errorf("GET %s: %v", link, err)
			failed++
			continue
		}
		linkResp.Body.Close()

		if linkResp.StatusCode >= 400 {
			t.Errorf("GET %s returned %d", link, linkResp.StatusCode)
			failed++
		} else {
			t.Logf("✓ %s (%d)", link, linkResp.StatusCode)
		}
	}

	if failed > 0 {
		t.Fatalf("%d/%d links failed", failed, len(links))
	}
}

func extractHrefs(resp *http.Response) []string {
	var links []string
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return links
	}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && attr.Val != "" && !strings.HasPrefix(attr.Val, "#") {
					links = append(links, attr.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	return links
}

func TestSheetMusicLinksDropbox(t *testing.T) {
	client := &http.Client{Timeout: 15 * time.Second}

	pageURL := baseURL() + "/sheet-music"
	resp, err := client.Get(pageURL)
	if err != nil {
		t.Fatalf("GET %s: %v", pageURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s returned %d", pageURL, resp.StatusCode)
	}

	links := extractHrefs(resp)
	var dropboxLinks []string
	for _, l := range links {
		if strings.Contains(l, "dropbox.com") {
			dropboxLinks = append(dropboxLinks, l)
		}
	}

	if len(dropboxLinks) == 0 {
		t.Fatal("no Dropbox links found on sheet music page")
	}

	t.Logf("found %d Dropbox links", len(dropboxLinks))

	failed := 0
	for _, link := range dropboxLinks {
		linkResp, err := client.Get(link)
		if err != nil {
			t.Errorf("GET %s: %v", link, err)
			failed++
			continue
		}
		linkResp.Body.Close()

		if linkResp.StatusCode >= 400 {
			t.Errorf("%s: %s", fmt.Sprintf("status %d", linkResp.StatusCode), link)
			failed++
		} else {
			t.Logf("✓ %s (%d)", link, linkResp.StatusCode)
		}
	}

	if failed > 0 {
		t.Fatalf("%d/%d Dropbox links failed", failed, len(dropboxLinks))
	}
}
