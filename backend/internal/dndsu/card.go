package dndsu

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// FetchCard downloads and parses a single monster card. edition must be
// "2014" or "2024" since the two sites use different markup for the
// statblock (confirmed live: 2014 uses .stat divs for abilities, 2024 uses
// .stat-pair-row divs with separate mod/save/roll columns).
func FetchCard(ctx context.Context, client *Client, url, edition string) (*ParsedCard, error) {
	html, err := client.Get(ctx, url)
	if err != nil {
		return nil, err
	}
	return ParseCard(html, url, edition)
}

// ParseCard parses a card's HTML. Exported for unit testing against fixtures.
func ParseCard(html, url, edition string) (*ParsedCard, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	card := &ParsedCard{
		Edition:   edition,
		SourceURL: url,
		RawHTML:   html,
	}
	card.DndsuID, card.Slug = idAndSlugFromURL(url)

	nameRu, nameEn := parseTitleNames(doc)
	card.NameRu, card.NameEn = nameRu, nameEn

	if img, ok := doc.Find(`meta[property="og:image"]`).Attr("content"); ok {
		card.ImageURL = img
	}

	header := doc.Find(".card__header h2.card-title")
	header.Find("span.source-plaque").First().Each(func(_ int, s *goquery.Selection) {
		if title, ok := s.Attr("title"); ok {
			card.SourceBook = title
		}
	})
	if otherHref, ok := header.Find("a.source-plaque").Attr("href"); ok {
		card.OtherEditionURL = otherHref
	}

	card.IsUniqueNPC = doc.Find(".list-icon__npc, [title='Именной НИП']").Length() > 0

	var parseErr error
	switch edition {
	case "2014":
		parseErr = parse2014Body(doc, card)
	case "2024":
		parseErr = parse2024Body(doc, card)
	default:
		return nil, fmt.Errorf("unknown edition %q", edition)
	}
	if parseErr != nil {
		card.ParseErrors = append(card.ParseErrors, parseErr.Error())
	}

	return card, nil
}

func idAndSlugFromURL(url string) (int, string) {
	trimmed := strings.Trim(url, "/")
	segments := strings.Split(trimmed, "/")
	if len(segments) == 0 {
		return 0, ""
	}
	last := segments[len(segments)-1]
	idPart := last
	if i := strings.Index(last, "-"); i > 0 {
		idPart = last[:i]
	}
	id, _ := strconv.Atoi(idPart)
	return id, last
}

// parseTitleNames extracts "NameRu [NameEn]" from the card title, e.g.
// "Ааракокра [Aarakocra]".
func parseTitleNames(doc *goquery.Document) (string, string) {
	raw := cleanText(doc.Find(".card__header h2.card-title span[data-copy]").First().Text())
	if raw == "" {
		raw = cleanText(doc.Find(".card__header h2.card-title").First().Text())
	}
	open := strings.Index(raw, "[")
	close := strings.LastIndex(raw, "]")
	if open > 0 && close > open {
		return cleanText(raw[:open]), cleanText(raw[open+1 : close])
	}
	return raw, ""
}
