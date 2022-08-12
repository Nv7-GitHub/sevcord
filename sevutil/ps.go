package sevutil

import (
	"fmt"
	"math"
	"strings"

	"github.com/Nv7-Github/sevcord"
)

type PageSwitchGetter struct {
	PageCount int
	Getter    func(page int) string
}

type PageSwitchData interface {
	[]string | PageSwitchGetter
}

type PageSwitcher[T PageSwitchData] struct {
	Title   string
	Content T

	// Optional
	Thumbnail  *string
	Footer     *string
	PageLength int // Default: 10
	Color      int

	// State
	page int
}

func (p *PageSwitcher[T]) build() *sevcord.Response {
	// Make content
	var content string
	var pages int
	switch cnt := any(p.Content).(type) {
	case []string:
		v := cnt[p.page*p.PageLength:]
		if len(v) > p.PageLength {
			v = v[:p.PageLength]
		}
		content = strings.Join(v, "\n")
		pages = int(math.Ceil(float64(len(cnt)) / float64(p.PageLength)))

	case PageSwitchGetter:
		content = cnt.Getter(p.page)
		pages = cnt.PageCount
	}

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

func NewPageSwitcher[T PageSwitchData](c sevcord.Ctx, p *PageSwitcher[T]) {
	if p.PageLength == 0 {
		p.PageLength = 10
	}
	c.Respond(p.build())
}
