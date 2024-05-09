package blog

import (
	"html/template"
	"os"
	"path/filepath"
	"runtime"

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

type Blog struct {
	Title       string
	Content     string
	Created     string
	URLVal      string
	FileName    string
	ContentHTML template.HTML
}

type BlogPageData struct {
	BlogPosts []Blog
}

var initializedBlogs = []Blog{}

var uninitializedBlogs = []Blog{
	{
		Title:    "Simple Docker Deploys",
		Created:  "May 8, 2024",
		FileName: "simpledockerdeploys.md",
		URLVal:   "simpledockerdeploys",
	},
	{
		Title:    "Thinking About What",
		Created:  "March 20, 2024",
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
			log.Info().Str("title", blog.Title).Str("content", blog.Content).Msg("found blog")
			return blog
		}
	}
	return Blog{}
}

// GetBlogs returns blog data for rendering in template
func GetBlogPageData() BlogPageData {
	return BlogPageData{BlogPosts: uninitializedBlogs}
}
