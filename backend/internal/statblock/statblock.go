// Package statblock defines the common contract shared by dnd.su monsters
// and user-created creatures, so both can be dropped into a combatant
// interchangeably.
package statblock

type Ability struct {
	Score int `json:"score"`
	Mod   int `json:"mod"`
}

type Abilities struct {
	Str Ability `json:"str"`
	Dex Ability `json:"dex"`
	Con Ability `json:"con"`
	Int Ability `json:"int"`
	Wis Ability `json:"wis"`
	Cha Ability `json:"cha"`
}

type Feature struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type StatBlock struct {
	SizeRu       string            `json:"sizeRu,omitempty"`
	Type         string            `json:"type,omitempty"`
	Alignment    string            `json:"alignment,omitempty"`
	ArmorClass   string            `json:"armorClass,omitempty"`
	ArmorSource  string            `json:"armorSource,omitempty"`
	Initiative   string            `json:"initiative,omitempty"`
	HitPoints    string            `json:"hitPoints,omitempty"`
	HitDice      string            `json:"hitDice,omitempty"`
	Speeds       map[string]string `json:"speeds,omitempty"`
	Abilities    Abilities         `json:"abilities"`
	SavingThrows map[string]string `json:"savingThrows,omitempty"`
	Skills       map[string]string `json:"skills,omitempty"`
	Vulnerabilities []string       `json:"vulnerabilities,omitempty"`
	Resistances     []string       `json:"resistances,omitempty"`
	Immunities      []string       `json:"immunities,omitempty"`
	ConditionImmunities []string   `json:"conditionImmunities,omitempty"`
	Senses             string      `json:"senses,omitempty"`
	PassivePerception  int         `json:"passivePerception,omitempty"`
	Languages          string      `json:"languages,omitempty"`
	ChallengeRating    string      `json:"challengeRating,omitempty"`
	ExperiencePoints   int         `json:"experiencePoints,omitempty"`
	ProficiencyBonus   int         `json:"proficiencyBonus,omitempty"`
	Traits         []Feature `json:"traits,omitempty"`
	Actions        []Feature `json:"actions,omitempty"`
	BonusActions   []Feature `json:"bonusActions,omitempty"`
	Reactions      []Feature `json:"reactions,omitempty"`
	LegendaryActions []Feature `json:"legendaryActions,omitempty"`
	Spellcasting   string `json:"spellcasting,omitempty"`
	LairActions    string `json:"lairActions,omitempty"`
	RegionalEffects string `json:"regionalEffects,omitempty"`
	SourceBook   string `json:"sourceBook,omitempty"`
	SourceURL    string `json:"sourceUrl,omitempty"`
	Habitat      string `json:"habitat,omitempty"`
}
