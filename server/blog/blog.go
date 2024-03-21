package blog

import "github.com/rs/zerolog/log"

type Blog struct {
	Title   string
	Content string
	Created string
}

type BlogPageData struct {
	BlogPosts []Blog
}

// GetBlogs returns a list of blogs
// TODO: This should read a group of markdown files
// from a directory and return them as a list of blogs
func GetBlogPageData() BlogPageData {
	log.Info().Msg("Getting blogs")
	blogs := []Blog{
		{
			Title:   "First Blog",
			Content: "This is the first blog",
			Created: "2021-01-01",
		},
		{
			Title:   "Second Blog",
			Content: "This is the second blog",
			Created: "2021-01-02",
		},
	}
	return BlogPageData{BlogPosts: blogs}
}
