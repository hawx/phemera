package markdown

import (
	"github.com/russross/blackfriday"
)

func Render(text string) string {
	flags := 0 |
		blackfriday.HTML_USE_XHTML |
		blackfriday.HTML_USE_SMARTYPANTS

	renderer := blackfriday.HtmlRenderer(flags, "", "")

	extensions := 0 |
		blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_SPACE_HEADERS

	return string(blackfriday.Markdown([]byte(text), renderer, extensions))
}
