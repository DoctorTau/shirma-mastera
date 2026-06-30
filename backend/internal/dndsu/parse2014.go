package dndsu

import (
	"strings"

	"github.com/PuerkitoBio/goquery"

	"monster-screen/backend/internal/statblock"
)

// parse2014Body parses the legacy dnd.su statblock markup:
// <ul class="params card__article-body"><li>...</li></ul>
func parse2014Body(doc *goquery.Document, card *ParsedCard) error {
	sb := &card.StatBlock
	params := doc.Find("ul.params").First()

	params.Children().Each(func(_ int, li *goquery.Selection) {
		class, _ := li.Attr("class")

		switch {
		case strings.Contains(class, "size-type-alignment"):
			parseSizeTypeAlignment(li, sb)
		case strings.Contains(class, "card-img__block"):
			// image already taken from og:image meta tag
		case strings.Contains(class, "abilities"):
			sb.Abilities, sb.SavingThrows = parseAbilities2014(li)
		case strings.Contains(class, "skills"):
			sb.Skills = parseSkills(li)
		case strings.Contains(class, "subsection") && strings.Contains(class, "desc"):
			parseSubsection2014(li, sb)
		default:
			label, value := liLabelValue(li)
			if label == "" {
				return
			}
			if strings.Contains(label, "Хиты") {
				sb.HitPoints, sb.HitDice = parseHP(li)
				return
			}
			applyCommonLabel(sb, label, value)
		}
	})

	sb.SourceBook = card.SourceBook
	sb.SourceURL = card.SourceURL
	return nil
}

var sizeRu = map[string]bool{
	"Крошечный": true, "Маленький": true, "Средний": true,
	"Большой": true, "Огромный": true, "Громадный": true,
}

func parseSizeTypeAlignment(li *goquery.Selection, sb *statblock.StatBlock) {
	clone := li.Clone()
	clone.Find("sup").Remove()
	text := cleanText(clone.Text())

	parts := strings.SplitN(text, ",", 2)
	head := cleanText(parts[0])
	if parts2 := strings.SplitN(head, " ", 2); len(parts2) == 2 {
		sb.SizeRu = parts2[0]
		sb.Type = cleanText(parts2[1])
	} else {
		sb.SizeRu = head
	}
	if len(parts) == 2 {
		sb.Alignment = cleanText(parts[1])
	}
}

func parseAbilities2014(li *goquery.Selection) (statblock.Abilities, map[string]string) {
	var abilities statblock.Abilities
	saves := map[string]string{}

	li.Find("div.stat").Each(func(_ int, stat *goquery.Selection) {
		title, _ := stat.Attr("title")
		valueText := cleanText(stat.Find("div").Eq(1).Text())
		ability := parseAbilityScore(valueText)

		switch title {
		case "Сила":
			abilities.Str = ability
		case "Ловкость":
			abilities.Dex = ability
		case "Телосложение":
			abilities.Con = ability
		case "Интеллект":
			abilities.Int = ability
		case "Мудрость":
			abilities.Wis = ability
		case "Харизма":
			abilities.Cha = ability
		}
	})

	return abilities, saves
}

func parseSubsection2014(li *goquery.Selection, sb *statblock.StatBlock) {
	if li.HasClass("additionalInfo") || li.Find(".additionalInfo").Length() > 0 {
		return
	}

	sectionTitle := cleanText(li.Find("h3.subsection-title").First().Text())

	body := li.Find("div").First()
	if body.Length() == 0 {
		body = li
	}

	var paragraphs []featureParagraph
	body.Find("p").Each(func(_ int, p *goquery.Selection) {
		paragraphs = append(paragraphs, parseFeatureParagraph2014(p))
	})
	if len(paragraphs) == 0 {
		return
	}

	bucketFeatures(sb, sectionTitle, paragraphs)
}

func parseFeatureParagraph2014(p *goquery.Selection) featureParagraph {
	nameNode := p.Find("strong").First()
	rawName := cleanText(nameNode.Text())
	name := strings.TrimSuffix(rawName, ".")

	fullText := cleanText(p.Text())
	desc := fullText
	if rawName != "" && strings.HasPrefix(fullText, rawName) {
		desc = cleanText(strings.TrimPrefix(fullText, rawName))
	}

	return featureParagraph{Name: name, Description: desc}
}
