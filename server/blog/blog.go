package blog

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gorilla/feeds"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/russross/blackfriday/v2"
)

func init() {
	var blogsposts []Blog
	for _, blog := range uninitializedBlogs {
		filepath := filepath.Join(basepath, "posts", blog.FileName)
		file, err := os.ReadFile(filepath)
		if err != nil {
			log.Error().Err(err).Msg("error reading blog file")
		}
		blogp := blog
		blogp.Content = string(file)
		blogsposts = append(blogsposts, blogp)
	}
	initializedBlogs = blogsposts
}

// HandleIndividualBlogPage handles returning the individual blog page
func HandleIndividualBlogPage(c echo.Context) error {
	requestedblog := GetBlog(c.Param("blog"))
	err := c.Render(http.StatusOK, "singleblogpage", requestedblog)
	if err != nil {
		return err
	}
	return nil
}

func HandleRssFeed(c echo.Context) error {
	now := time.Now()
	feed := &feeds.Feed{
		Title:       "Andrew Willette's Blog",
		Link:        &feeds.Link{Href: "https://andrewwillette.com/blog"},
		Description: "Latest updates from my blog.",
		Created:     now,
	}
	var feedItems []*feeds.Item
	for _, blog := range initializedBlogs {
		created, err := time.Parse("January 2, 2006", blog.Created)
		if err != nil {
			log.Error().Err(err).Msg("error parsing blog created date")
		}
		feedItems = append(feedItems, &feeds.Item{
			Title:   blog.Title,
			Link:    &feeds.Link{Href: "https://andrewwillette.com/blog/" + blog.URLVal},
			Content: blog.Content,
			Created: created,
			Updated: created,
		})
	}
	feed.Items = feedItems
	rss, err := feed.ToRss()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Unable to generate RSS feed")
	}
	return c.Blob(http.StatusOK, "application/rss+xml", []byte(rss))
}

// HandleBlogPage handles returning the blog page displaying all blogs
func HandleBlogPage(c echo.Context) error {
	err := c.Render(http.StatusOK, "blogspage", GetBlogPageData())
	if err != nil {
		return err
	}
	return nil
}

type Blog struct {
	Title       string
	Content     string
	Created     string
	URLVal      string
	FileName    string
	ContentHTML template.HTML
	CurrentYear int
}

type BlogPageData struct {
	BlogPosts   []Blog
	CurrentYear int
}

var initializedBlogs = []Blog{}

var uninitializedBlogs = []Blog{
	{
		Title:    "Discipline In 2025 USA",
		Created:  time.Date(2025, time.September, 4, 0, 0, 0, 0, time.UTC).Format("January 2, 2006"),
		FileName: "discipline.md",
		URLVal:   "theneedfordiscipline",
	},
	{
		Title:    "Key of the Day",
		Created:  time.Date(2024, time.November, 24, 0, 0, 0, 0, time.UTC).Format("January 2, 2006"),
		FileName: "keyoftheday.md",
		URLVal:   "keyoftheday",
	},
	{
		Title:    "Simple Docker Deploys",
		Created:  time.Date(2024, time.May, 8, 0, 0, 0, 0, time.UTC).Format("January 2, 2006"),
		FileName: "simpledockerdeploys.md",
		URLVal:   "simpledockerdeploys",
	},
	{
		Title:    "Thinking About What",
		Created:  time.Date(2024, time.March, 20, 0, 0, 0, 0, time.UTC).Format("January 2, 2006"),
		FileName: "thinkingaboutwhat.md",
		URLVal:   "thinkingaboutwhat",
	},
}
var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

func GetBlog(urlval string) Blog {
	for _, blog := range initializedBlogs {
		if blog.URLVal == urlval {
			output := blackfriday.Run([]byte(blog.Content), blackfriday.WithExtensions(blackfriday.CommonExtensions))
			blog.Content = string(output)
			blog.ContentHTML = template.HTML(blog.Content)
			blog.CurrentYear = time.Now().Year()
			return blog
		}
	}
	return Blog{}
}

// GetBlogs returns blog data for rendering in template
func GetBlogPageData() BlogPageData {
	currentYear := time.Now().Year()
	return BlogPageData{
		BlogPosts:   uninitializedBlogs,
		CurrentYear: currentYear,
	}
}
