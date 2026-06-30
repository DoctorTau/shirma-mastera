package dndsu

import (
	"context"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type CatalogEntry struct {
	DndsuID     int
	Slug        string
	URL         string
	NameRu      string
	NameEn      string
	CR          string
	IsUniqueNPC bool
}

// CatalogIndexListPath is the AJAX data endpoint the dnd.su /bestiary/ page
// fetches client-side instead of rendering the list server-side. Confirmed
// live: returns the full 2014-edition catalog (~2900 entries) regardless of
// the `content` query param value.
const CatalogIndexListPath = "/piece/bestiary/index-list/?content=multiverse"

// FetchCatalog2014 walks the single AJAX list endpoint that backs dnd.su's
// bestiary page and returns every monster entry found in it.
func FetchCatalog2014(ctx context.Context, client *Client, baseURL string) ([]CatalogEntry, error) {
	html, err := client.Get(ctx, baseURL+CatalogIndexListPath)
	if err != nil {
		return nil, err
	}
	return ParseCatalogFragment(html, baseURL)
}

// ParseCatalogFragment parses the HTML fragment served by the index-list
// endpoint. Exported for unit testing against a saved fixture.
func ParseCatalogFragment(html string, baseURL string) ([]CatalogEntry, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	var entries []CatalogEntry
	doc.Find("div.list-item__beast").Each(func(_ int, sel *goquery.Selection) {
		idStr, _ := sel.Attr("data-id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return
		}

		href, _ := sel.Find("a.list-item-wrapper").Attr("href")
		slug := slugFromURL(href)

		search, _ := sel.Attr("data-search")
		nameRu, nameEn := splitSearchNames(search)
		if nameRu == "" {
			nameRu = strings.TrimSpace(sel.Find(".list-item-title").Text())
		}

		cr := strings.TrimSpace(sel.Find(".list-mark__danger span").First().Text())
		isUnique := sel.Find(".list-icon__npc").Length() > 0

		entries = append(entries, CatalogEntry{
			DndsuID:     id,
			Slug:        slug,
			URL:         resolveURL(baseURL, href),
			NameRu:      nameRu,
			NameEn:      nameEn,
			CR:          cr,
			IsUniqueNPC: isUnique,
		})
	})

	return entries, nil
}

// splitSearchNames parses dnd.su's `data-search='NameRu,NameEn,'` attribute.
func splitSearchNames(search string) (nameRu, nameEn string) {
	parts := strings.Split(search, ",")
	if len(parts) > 0 {
		nameRu = strings.TrimSpace(parts[0])
	}
	if len(parts) > 1 {
		nameEn = strings.TrimSpace(parts[1])
	}
	return
}

// slugFromURL extracts "30-aarakocra" style slugs (with the leading id) from
// a card href such as "/bestiary/30-aarakocra/".
func slugFromURL(href string) string {
	href = strings.Trim(href, "/")
	segments := strings.Split(href, "/")
	if len(segments) == 0 {
		return ""
	}
	return segments[len(segments)-1]
}

func resolveURL(baseURL, href string) string {
	if strings.HasPrefix(href, "http") {
		return href
	}
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(href, "/")
}
