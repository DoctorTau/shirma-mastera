package dndsu

import (
	"regexp"
	"strconv"
	"strings"

	"monster-screen/backend/internal/statblock"
)

// ParsedCard is the normalized result of parsing a single dnd.su / next.dnd.su
// monster card. Field-level parse failures are recorded in ParseErrors
// instead of aborting the whole card — see the spec's §9.3 requirement that
// the parser must "не падать" (not fail) on a single bad field.
type ParsedCard struct {
	DndsuID         int
	Slug            string
	NameRu          string
	NameEn          string
	Edition         string
	SourceBook      string
	SourceURL       string
	ImageURL        string
	StatBlock       statblock.StatBlock
	OtherEditionURL string
	IsUniqueNPC     bool
	RawHTML         string
	ParseErrors     []string
}

var abilityScoreRe = regexp.MustCompile(`(-?\d+)\s*\(([+-]\d+)\)`)
var signedNumberRe = regexp.MustCompile(`([+-]?\d+)`)
var crRe = regexp.MustCompile(`^\s*([\d/]+)`)
var xpRe = regexp.MustCompile(`([\d\s]+)\s*опыта`)

func parseAbilityScore(text string) statblock.Ability {
	m := abilityScoreRe.FindStringSubmatch(text)
	if len(m) != 3 {
		return statblock.Ability{}
	}
	score, _ := strconv.Atoi(m[1])
	mod, _ := strconv.Atoi(m[2])
	return statblock.Ability{Score: score, Mod: mod}
}

func parseFirstSignedNumber(text string) (int, bool) {
	m := signedNumberRe.FindString(text)
	if m == "" {
		return 0, false
	}
	n, err := strconv.Atoi(m)
	if err != nil {
		return 0, false
	}
	return n, true
}

// CRToNumeric converts a challenge rating string like "1/4", "1/2", "0" or
// "13" into a sortable/filterable decimal.
func CRToNumeric(cr string) *float64 {
	m := crRe.FindStringSubmatch(cr)
	if len(m) != 2 {
		return nil
	}
	val := m[1]
	if strings.Contains(val, "/") {
		parts := strings.SplitN(val, "/", 2)
		num, err1 := strconv.ParseFloat(parts[0], 64)
		den, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 != nil || err2 != nil || den == 0 {
			return nil
		}
		v := num / den
		return &v
	}
	v, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return nil
	}
	return &v
}

func parseXP(text string) int {
	m := xpRe.FindStringSubmatch(text)
	if len(m) != 2 {
		return 0
	}
	digits := strings.ReplaceAll(m[1], " ", "")
	n, _ := strconv.Atoi(digits)
	return n
}

func cleanText(s string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

// stripLabel removes a leading "<strong>Label</strong>" worth of text once
// the caller already has the li's full text; used after deleting the
// <strong> node from the DOM selection so only the value text remains.
func stripLabel(s string) string {
	return cleanText(s)
}
