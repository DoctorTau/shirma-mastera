package dndsu

import (
	"os"
	"testing"
)

func readFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return string(data)
}

func TestParseCard_Aarakocra2014(t *testing.T) {
	html := readFixture(t, "aarakocra_2014.html")
	card, err := ParseCard(html, "https://dnd.su/bestiary/30-aarakocra/", "2014")
	if err != nil {
		t.Fatalf("ParseCard: %v", err)
	}

	if card.NameRu != "Ааракокра" {
		t.Errorf("NameRu = %q, want Ааракокра", card.NameRu)
	}
	if card.NameEn != "Aarakocra" {
		t.Errorf("NameEn = %q, want Aarakocra", card.NameEn)
	}
	if card.DndsuID != 30 {
		t.Errorf("DndsuID = %d, want 30", card.DndsuID)
	}
	if card.Slug != "30-aarakocra" {
		t.Errorf("Slug = %q, want 30-aarakocra", card.Slug)
	}
	if card.SourceBook != "Monster Manual" {
		t.Errorf("SourceBook = %q, want Monster Manual", card.SourceBook)
	}
	if card.OtherEditionURL != "https://next.dnd.su/bestiary/21153-aarakocra" {
		t.Errorf("OtherEditionURL = %q, want the 2024 cross-link", card.OtherEditionURL)
	}
	if card.ImageURL == "" {
		t.Error("ImageURL is empty")
	}

	sb := card.StatBlock
	if sb.SizeRu != "Средний" {
		t.Errorf("SizeRu = %q, want Средний", sb.SizeRu)
	}
	if sb.Type != "Гуманоид" {
		t.Errorf("Type = %q, want Гуманоид", sb.Type)
	}
	if sb.Alignment != "нейтрально-добрый" {
		t.Errorf("Alignment = %q, want нейтрально-добрый", sb.Alignment)
	}
	if sb.ArmorClass != "12" {
		t.Errorf("ArmorClass = %q, want 12", sb.ArmorClass)
	}
	if sb.HitPoints != "13" {
		t.Errorf("HitPoints = %q, want 13", sb.HitPoints)
	}
	if sb.HitDice != "3к8" {
		t.Errorf("HitDice = %q, want 3к8", sb.HitDice)
	}
	if sb.Speeds["ходьба"] != "20 футов" {
		t.Errorf("Speeds[ходьба] = %q, want 20 футов", sb.Speeds["ходьба"])
	}
	if sb.Speeds["летая"] != "50 футов" {
		t.Errorf("Speeds[летая] = %q, want 50 футов", sb.Speeds["летая"])
	}
	if sb.Abilities.Dex.Score != 14 || sb.Abilities.Dex.Mod != 2 {
		t.Errorf("Abilities.Dex = %+v, want score=14 mod=2", sb.Abilities.Dex)
	}
	if sb.Abilities.Str.Score != 10 || sb.Abilities.Str.Mod != 0 {
		t.Errorf("Abilities.Str = %+v, want score=10 mod=0", sb.Abilities.Str)
	}
	if sb.Skills["Восприятие"] != "+5" {
		t.Errorf("Skills[Восприятие] = %q, want +5", sb.Skills["Восприятие"])
	}
	if sb.PassivePerception != 15 {
		t.Errorf("PassivePerception = %d, want 15", sb.PassivePerception)
	}
	if sb.Languages != "Ауран" {
		t.Errorf("Languages = %q, want Ауран", sb.Languages)
	}
	if sb.ChallengeRating != "1/4" {
		t.Errorf("ChallengeRating = %q, want 1/4", sb.ChallengeRating)
	}
	if sb.ExperiencePoints != 50 {
		t.Errorf("ExperiencePoints = %d, want 50", sb.ExperiencePoints)
	}
	if sb.ProficiencyBonus != 2 {
		t.Errorf("ProficiencyBonus = %d, want 2", sb.ProficiencyBonus)
	}
	if len(sb.Traits) != 1 || sb.Traits[0].Name != "Пикирующая атака" {
		t.Errorf("Traits = %+v, want one trait named Пикирующая атака", sb.Traits)
	}
	if len(sb.Actions) != 3 {
		t.Fatalf("len(Actions) = %d, want 3", len(sb.Actions))
	}
	if sb.Actions[0].Name != "Коготь" {
		t.Errorf("Actions[0].Name = %q, want Коготь", sb.Actions[0].Name)
	}
	if sb.Actions[0].Description == "" {
		t.Error("Actions[0].Description is empty")
	}
}

func TestParseCard_Aarakocra2024(t *testing.T) {
	html := readFixture(t, "aarakocra_2024.html")
	card, err := ParseCard(html, "https://next.dnd.su/bestiary/21153-aarakocra-skirmisher/", "2024")
	if err != nil {
		t.Fatalf("ParseCard: %v", err)
	}

	if card.NameRu != "Ааракокра застрельщик" {
		t.Errorf("NameRu = %q, want Ааракокра застрельщик", card.NameRu)
	}
	if card.OtherEditionURL != "https://5e14.dnd.su/bestiary/30-aarakocra" {
		t.Errorf("OtherEditionURL = %q, want the 2014 cross-link", card.OtherEditionURL)
	}

	sb := card.StatBlock
	if sb.ArmorClass != "12" {
		t.Errorf("ArmorClass = %q, want 12", sb.ArmorClass)
	}
	if sb.Initiative == "" {
		t.Error("Initiative is empty, want a parsed value like +2 (12)")
	}
	if sb.Abilities.Dex.Score != 14 || sb.Abilities.Dex.Mod != 2 {
		t.Errorf("Abilities.Dex = %+v, want score=14 mod=2", sb.Abilities.Dex)
	}
	if sb.HitDice == "" {
		t.Error("HitDice is empty")
	}
	if len(sb.Actions) == 0 {
		t.Error("Actions is empty, want at least one parsed action")
	}
	for _, a := range sb.Actions {
		if a.Name == "" {
			t.Errorf("action with empty name: %+v", a)
		}
	}
}

func TestParseCard_Behir2014DamageImmunity(t *testing.T) {
	html := readFixture(t, "behir_2014.html")
	card, err := ParseCard(html, "https://dnd.su/bestiary/42-behir/", "2014")
	if err != nil {
		t.Fatalf("ParseCard: %v", err)
	}

	sb := card.StatBlock
	if len(sb.Immunities) != 1 || sb.Immunities[0] != "электричество" {
		t.Errorf("Immunities = %+v, want [электричество]", sb.Immunities)
	}
	if sb.ChallengeRating != "11" {
		t.Errorf("ChallengeRating = %q, want 11", sb.ChallengeRating)
	}
	if len(sb.Actions) == 0 {
		t.Error("Actions is empty")
	}
	if sb.Actions[0].Name != "Мультиатака" {
		t.Errorf("Actions[0].Name = %q, want Мультиатака", sb.Actions[0].Name)
	}
}

func TestParseCard_Behir2024(t *testing.T) {
	html := readFixture(t, "behir_2024.html")
	card, err := ParseCard(html, "https://next.dnd.su/bestiary/21211-behir/", "2024")
	if err != nil {
		t.Fatalf("ParseCard: %v", err)
	}

	sb := card.StatBlock
	if len(sb.Immunities) == 0 {
		t.Error("Immunities is empty, want at least electricity")
	}
	if len(sb.Actions) == 0 {
		t.Error("Actions is empty")
	}
	foundMultiattack := false
	for _, a := range sb.Actions {
		if a.Name == "Мультиатака" {
			foundMultiattack = true
		}
	}
	if !foundMultiattack {
		t.Errorf("Actions = %+v, want a Мультиатака entry", sb.Actions)
	}
}

func TestParseCard_NoOtherEdition(t *testing.T) {
	html := readFixture(t, "alastrah_2014_no2024.html")
	card, err := ParseCard(html, "https://dnd.su/bestiary/7758-alastrah/", "2014")
	if err != nil {
		t.Fatalf("ParseCard: %v", err)
	}
	if card.OtherEditionURL != "" {
		t.Errorf("OtherEditionURL = %q, want empty (this monster has no 2024 version)", card.OtherEditionURL)
	}
	if card.SourceBook != "Storm King's Thunder" {
		t.Errorf("SourceBook = %q, want Storm King's Thunder", card.SourceBook)
	}
}
