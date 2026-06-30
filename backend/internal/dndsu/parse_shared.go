package dndsu

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"monster-screen/backend/internal/statblock"
)

// liLabelValue returns the cleaned label (first <strong>) and the remaining
// text of a "<strong>Label</strong> value" style <li>, with tooltip <sup>
// nodes stripped first so they don't leak a stray "?" into the value.
func liLabelValue(li *goquery.Selection) (label, value string) {
	clone := li.Clone()
	clone.Find("sup").Remove()
	label = cleanText(clone.Find("strong").First().Text())
	clone.Find("strong").First().Remove()
	value = cleanText(clone.Text())
	return
}

var acValueRe = regexp.MustCompile(`^(\d+)\s*(?:\(([^)]*)\))?`)

func parseACValue(value string) (ac, source string) {
	m := acValueRe.FindStringSubmatch(value)
	if len(m) == 0 {
		return cleanText(value), ""
	}
	return m[1], cleanText(m[2])
}

var speedTypeAndValueRe = regexp.MustCompile(`^(.*?)\s*(\d+\s*фут[\p{L}]*\.?)\s*$`)

// parseSpeeds turns "20 футов, летая 50 футов" into
// {"ходьба": "20 футов", "летая": "50 футов"}.
func parseSpeeds(value string) map[string]string {
	speeds := map[string]string{}
	for _, chunk := range strings.Split(value, ",") {
		chunk = cleanText(chunk)
		if chunk == "" {
			continue
		}
		m := speedTypeAndValueRe.FindStringSubmatch(chunk)
		if len(m) != 3 {
			speeds["ходьба"] = chunk
			continue
		}
		kind := cleanText(m[1])
		if kind == "" {
			kind = "ходьба"
		}
		speeds[kind] = cleanText(m[2])
	}
	return speeds
}

func splitList(value string) []string {
	if cleanText(value) == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(value, ",") {
		part = cleanText(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

var passivePerceptionRe = regexp.MustCompile(`[Пп]ассивное\s+Восприятие\s*(\d+)`)

func parsePassivePerception(value string) (int, bool) {
	m := passivePerceptionRe.FindStringSubmatch(value)
	if len(m) != 2 {
		return 0, false
	}
	n, err := strconv.Atoi(m[1])
	return n, err == nil
}

var proficiencyInCRRe = regexp.MustCompile(`БВ\s*([+-]?\d+)`)

func parseProficiencyFromCR(value string) (int, bool) {
	m := proficiencyInCRRe.FindStringSubmatch(value)
	if len(m) != 2 {
		return 0, false
	}
	n, err := strconv.Atoi(m[1])
	return n, err == nil
}

func crPart(value string) string {
	m := crRe.FindStringSubmatch(value)
	if len(m) != 2 {
		return cleanText(value)
	}
	return m[1]
}

// applyCommonLabel maps a label/value pair to the shared StatBlock fields
// that are spelled (almost) identically across the 2014 and 2024 markup.
// Returns false if the label wasn't recognized so the caller can decide
// what to do with leftovers instead of silently dropping them.
func applyCommonLabel(sb *statblock.StatBlock, label, value string) bool {
	switch {
	case strings.Contains(label, "Класс Доспеха"), strings.Contains(label, "Класс Защиты"):
		sb.ArmorClass, sb.ArmorSource = parseACValue(value)
	case strings.Contains(label, "Скорость"):
		sb.Speeds = parseSpeeds(value)
	case strings.Contains(label, "Уязвимост"):
		sb.Vulnerabilities = splitList(value)
	case strings.Contains(label, "Сопротивл"):
		sb.Resistances = splitList(value)
	case strings.Contains(label, "Иммунитет") && strings.Contains(label, "состоян"):
		sb.ConditionImmunities = splitList(value)
	case strings.Contains(label, "Иммунитет"):
		sb.Immunities = splitList(value)
	case strings.Contains(label, "Чувства"):
		sb.Senses = value
		if pp, ok := parsePassivePerception(value); ok {
			sb.PassivePerception = pp
		}
	case strings.Contains(label, "Языки"):
		sb.Languages = value
	case strings.Contains(label, "Опасность"):
		sb.ChallengeRating = crPart(value)
		sb.ExperiencePoints = parseXP(value)
		if pb, ok := parseProficiencyFromCR(value); ok {
			sb.ProficiencyBonus = pb
		}
	case strings.Contains(label, "Бонус мастерства"):
		if n, ok := parseFirstSignedNumber(label); ok {
			sb.ProficiencyBonus = n
		}
	case strings.Contains(label, "Местность обитания"), strings.Contains(label, "Среда обитания"):
		sb.Habitat = value
	case strings.Contains(label, "Спасброски"):
		sb.SavingThrows = parseInlineBonusList(value)
	default:
		return false
	}
	return true
}

// parseInlineBonusList turns "Лов +5, Тел +4" into {"Лов": "+5", "Тел": "+4"}.
func parseInlineBonusList(value string) map[string]string {
	out := map[string]string{}
	for _, part := range strings.Split(value, ",") {
		part = cleanText(part)
		idx := strings.LastIndex(part, " ")
		if idx <= 0 {
			continue
		}
		name := cleanText(part[:idx])
		bonus := cleanText(part[idx+1:])
		if name != "" && bonus != "" {
			out[name] = bonus
		}
	}
	return out
}

// parseSkills reads the dnd.su skills markup, identical on both editions:
// <span class='skill-bonus'>Name <strong class='skill-bonus-value'>+N</strong></span>
func parseSkills(li *goquery.Selection) map[string]string {
	skills := map[string]string{}
	li.Find("span.skill-bonus").Each(func(_ int, s *goquery.Selection) {
		bonus := cleanText(s.Find(".skill-bonus-value").Text())
		clone := s.Clone()
		clone.Find(".skill-bonus-value").Remove()
		name := cleanText(clone.Text())
		if name != "" && bonus != "" {
			skills[name] = bonus
		}
	})
	return skills
}

// parseHP reads the dnd.su hit point markup, identical on both editions:
// <span data-type='middle'>N</span> (<span data-type='throw'>X</span>к<span data-type='dice'>Y</span>
// [<span data-type='action'>+</span><span data-type='bonus'>Z</span>])
func parseHP(li *goquery.Selection) (average, formula string) {
	average = cleanText(li.Find("span[data-type='middle']").First().Text())
	throw := cleanText(li.Find("span[data-type='throw']").First().Text())
	dice := cleanText(li.Find("span[data-type='dice']").First().Text())
	action := cleanText(li.Find("span[data-type='action']").First().Text())
	bonus := cleanText(li.Find("span[data-type='bonus']").First().Text())

	if throw == "" || dice == "" {
		return average, ""
	}
	formula = throw + "к" + dice
	if action != "" && bonus != "" {
		formula += action + bonus
	}
	return average, formula
}

type featureParagraph struct {
	Name        string
	Description string
}

// bucketFeatures appends parsed paragraphs into the right StatBlock list (or
// raw text block) based on the section heading text. Unrecognized headings
// default to Traits rather than being dropped, matching the spec's "store
// what parsed, never abort" requirement.
func bucketFeatures(sb *statblock.StatBlock, sectionTitle string, paragraphs []featureParagraph) {
	title := sectionTitle
	asFeatures := func() []statblock.Feature {
		out := make([]statblock.Feature, 0, len(paragraphs))
		for _, p := range paragraphs {
			out = append(out, statblock.Feature{Name: p.Name, Description: p.Description})
		}
		return out
	}
	asText := func() string {
		var sb strings.Builder
		for i, p := range paragraphs {
			if i > 0 {
				sb.WriteString("\n\n")
			}
			if p.Name != "" {
				sb.WriteString(p.Name + ". ")
			}
			sb.WriteString(p.Description)
		}
		return sb.String()
	}

	switch {
	case title == "":
		sb.Traits = append(sb.Traits, asFeatures()...)
	case strings.Contains(title, "Описание"), strings.Contains(title, "Игровой персонаж"):
		// flavor text, not part of the v1 statblock contract
		return
	case strings.Contains(title, "Бонусные действия"):
		sb.BonusActions = append(sb.BonusActions, asFeatures()...)
	case strings.Contains(title, "Реакции"):
		sb.Reactions = append(sb.Reactions, asFeatures()...)
	case strings.Contains(title, "Легендарные действия"):
		sb.LegendaryActions = append(sb.LegendaryActions, asFeatures()...)
	case strings.Contains(title, "Действия"):
		sb.Actions = append(sb.Actions, asFeatures()...)
	case strings.Contains(title, "Логово"):
		sb.LairActions = asText()
	case strings.Contains(title, "Региональ"):
		sb.RegionalEffects = asText()
	case strings.Contains(title, "Закл"):
		sb.Spellcasting = asText()
	default:
		sb.Traits = append(sb.Traits, asFeatures()...)
	}
}
