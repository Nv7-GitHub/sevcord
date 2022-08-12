package sevutil

import (
	"fmt"
	"math"
	"strings"

	"github.com/Nv7-Github/sevcord"
)

func PSGetterFromItems[T any](items []T, pageLength int) PageSwitchGetter {
	vals := make([]string, len(items))
	for i, item := range items {
		vals[i] = fmt.Sprintf("%v", item)
	}
	return PageSwitchGetter{
		PageCount: int(math.Ceil(float64(len(items)) / float64(pageLength))),
		Getter: func(page int) string {
			v := vals[page*pageLength:]
			if len(v) > pageLength {
				v = v[:pageLength]
			}
			return strings.Join(v, "\n")
		},
	}
}

type PageSwitchGetter struct {
	PageCount int
	Getter    func(page int) string
}

type PageSwitcher struct {
	Title   string
	Content PageSwitchGetter

	// Optional
	Thumbnail *string
	Footer    *string
	Color     int

	// State
	page int
}

func (p *PageSwitcher) build() *sevcord.Response {
	// Make content
	content := p.Content.Getter(p.page)
	pages := p.Content.PageCount

	// Build embed
	bld := sevcord.NewEmbedBuilder(p.Title).Color(p.Color).Description(content)
	if p.Thumbnail != nil {
		bld.Thumbnail(*p.Thumbnail)
	}
	footer := fmt.Sprintf("Page %d/%d", p.page+1, pages)
	if p.Footer != nil {
		footer += " • " + *p.Footer
	}
	bld.Footer(footer, "")

	// Add btns
	resp := sevcord.EmbedResponse(bld)
	resp.ComponentRow(&sevcord.Button{
		Style: sevcord.ButtonStylePrimary,
		Emoji: sevcord.ComponentEmojiCustom("leftarrow", "861722690813165598", false),
		Handler: func(c sevcord.Ctx) {
			p.page -= 1
			if p.page < 0 {
				p.page = pages - 1 // Loop Around
			}
			c.Edit(p.build())
		},
	}, &sevcord.Button{
		Style: sevcord.ButtonStylePrimary,
		Emoji: sevcord.ComponentEmojiCustom("rightarrow", "861722690926936084", false),
		Handler: func(c sevcord.Ctx) {
			p.page += 1
			if p.page >= pages {
				p.page = 0 // Loop Around
			}
			c.Edit(p.build())
		},
	})

	return resp
}

func NewPageSwitcher(c sevcord.Ctx, p *PageSwitcher) {
	c.Respond(p.build())
}
