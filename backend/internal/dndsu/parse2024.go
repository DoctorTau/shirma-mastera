package dndsu

import (
	"strings"

	"github.com/PuerkitoBio/goquery"

	"monster-screen/backend/internal/statblock"
)

// parse2024Body parses the next.dnd.su statblock markup, which is
// structurally different from 2014's: AC+Initiative share one <li>,
// abilities use stat-pair-row divs (mod/save/roll columns), and
// traits/actions use article-body__monster__section instead of
// subsection/desc.
func parse2024Body(doc *goquery.Document, card *ParsedCard) error {
	sb := &card.StatBlock
	params := doc.Find("ul.params").First()

	params.Children().Each(func(_ int, li *goquery.Selection) {
		class, _ := li.Attr("class")

		switch {
		case strings.Contains(class, "size-type-alignment"):
			parseSizeTypeAlignment(li, sb)
		case strings.Contains(class, "card-img__block"):
			// image already taken from og:image meta tag
		case strings.Contains(class, "armor-class-and-initiative"):
			parseACAndInitiative2024(li, sb)
		case strings.Contains(class, "abilities"):
			sb.Abilities, sb.SavingThrows = parseAbilities2024(li)
		case strings.Contains(class, "skills"):
			sb.Skills = parseSkills(li)
		case strings.Contains(class, "monster__section"):
			parseSection2024(li, sb)
		default:
			label, value := liLabelValue(li)
			if label == "" {
				return
			}
			if strings.Contains(label, "Хиты") {
				sb.HitPoints, sb.HitDice = parseHP(li)
				return
			}
			if strings.Contains(label, "Сокровища") {
				return
			}
			applyCommonLabel(sb, label, value)
		}
	})

	sb.SourceBook = card.SourceBook
	sb.SourceURL = card.SourceURL
	return nil
}

func parseACAndInitiative2024(li *goquery.Selection, sb *statblock.StatBlock) {
	ac := li.Find(".subsection-ac").First()
	_, acValue := liLabelValue(ac)
	sb.ArmorClass, sb.ArmorSource = parseACValue(acValue)

	initiative := li.Find(".subsection-initiative").First()
	_, initValue := liLabelValue(initiative)
	sb.Initiative = cleanText(initValue)
}

var abilityAbbrevRu = map[string]bool{"СИЛ": true, "ЛОВ": true, "ТЕЛ": true, "ИНТ": true, "МДР": true, "ХАР": true}

func parseAbilities2024(li *goquery.Selection) (statblock.Abilities, map[string]string) {
	var abilities statblock.Abilities
	saves := map[string]string{}

	li.Find("div.stat-pair-row").Each(func(_ int, row *goquery.Selection) {
		title := cleanText(row.Find(".title").Text())
		scoreText := cleanText(row.Find(".value").Text())
		modText := cleanText(row.Find(".mod").Text())
		saveText := cleanText(row.Find(".save").Text())

		score, _ := parseFirstSignedNumber(scoreText)
		mod, _ := parseFirstSignedNumber(modText)
		ability := statblock.Ability{Score: score, Mod: mod}

		if saveText != "" && saveText != modText {
			saves[title] = saveText
		}

		switch title {
		case "СИЛ":
			abilities.Str = ability
		case "ЛОВ":
			abilities.Dex = ability
		case "ТЕЛ":
			abilities.Con = ability
		case "ИНТ":
			abilities.Int = ability
		case "МДР":
			abilities.Wis = ability
		case "ХАР":
			abilities.Cha = ability
		}
	})

	return abilities, saves
}

func parseSection2024(li *goquery.Selection, sb *statblock.StatBlock) {
	sectionTitle := cleanText(li.Find("h3").First().Text())
	body := li.Find(".article-body__monster__section-body").First()
	if body.Length() == 0 {
		body = li
	}

	var paragraphs []featureParagraph
	body.Find("p").Each(func(_ int, p *goquery.Selection) {
		paragraphs = append(paragraphs, parseFeatureParagraph2024(p))
	})
	if len(paragraphs) == 0 {
		return
	}
	bucketFeatures(sb, sectionTitle, paragraphs)
}

func parseFeatureParagraph2024(p *goquery.Selection) featureParagraph {
	nameNode := p.Find(".article-body__feature-name, .article-body__monster-attack__name").First()
	rawName := cleanText(nameNode.Text())
	name := strings.TrimSuffix(rawName, ".")

	fullText := cleanText(p.Text())
	desc := fullText
	if rawName != "" {
		prefix := rawName
		if !strings.HasSuffix(prefix, ".") {
			prefix += "."
		}
		if strings.HasPrefix(fullText, prefix) {
			desc = cleanText(strings.TrimPrefix(fullText, prefix))
		} else if strings.HasPrefix(fullText, rawName) {
			desc = cleanText(strings.TrimPrefix(fullText, rawName))
		}
	}

	return featureParagraph{Name: name, Description: desc}
}
