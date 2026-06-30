export interface Ability {
	score: number;
	mod: number;
}

export interface Abilities {
	str: Ability;
	dex: Ability;
	con: Ability;
	int: Ability;
	wis: Ability;
	cha: Ability;
}

export interface Feature {
	name: string;
	description: string;
}

// Common contract: shared by dnd.su monsters and user-created creatures so
// either can be dropped into a combatant interchangeably.
export interface StatBlock {
	sizeRu?: string;
	type?: string;
	alignment?: string;
	armorClass?: string;
	armorSource?: string;
	initiative?: string;
	hitPoints?: string;
	hitDice?: string;
	speeds?: Record<string, string>;
	abilities: Abilities;
	savingThrows?: Record<string, string>;
	skills?: Record<string, string>;
	vulnerabilities?: string[];
	resistances?: string[];
	immunities?: string[];
	conditionImmunities?: string[];
	senses?: string;
	passivePerception?: number;
	languages?: string;
	challengeRating?: string;
	experiencePoints?: number;
	proficiencyBonus?: number;
	traits?: Feature[];
	actions?: Feature[];
	bonusActions?: Feature[];
	reactions?: Feature[];
	legendaryActions?: Feature[];
	spellcasting?: string;
	lairActions?: string;
	regionalEffects?: string;
	sourceBook?: string;
	sourceUrl?: string;
	habitat?: string;
}

export type Edition = '2014' | '2024';

export interface Monster {
	id: string;
	dndsuId: number;
	slug: string;
	edition: Edition;
	nameRu: string;
	nameEn: string;
	cr: string;
	type: string;
	size: string;
	alignment: string;
	statblock: StatBlock;
	imageUrl?: string;
	sourceBook?: string;
	sourceUrl: string;
	linkedMonsterId?: string;
	isUniqueNpc: boolean;
	lastFetchedAt?: string;
	updatedAt: string;
}

export interface CreatedCreature {
	id: string;
	nameRu: string;
	nameEn: string;
	statblock: StatBlock;
	notes: string;
	createdAt: string;
	updatedAt: string;
}

export interface PlayerCharacter {
	id: string;
	name: string;
	ac?: number;
	passivePerception?: number;
	maxHp?: number;
	notes: string;
	createdAt: string;
	updatedAt: string;
}

export type CombatantSourceType = 'monster' | 'created_creature' | 'player_character' | 'custom';

export interface Combatant {
	id: string;
	encounterId: string;
	sourceType: CombatantSourceType;
	sourceId?: string;
	monsterEdition?: Edition;
	displayName: string;
	maxHp?: number;
	currentHp?: number;
	tempHp: number;
	initiative?: number;
	conditions: string[];
	notes: string;
	isPc: boolean;
	sortOrder: number;
	updatedAt: string;
}

export interface Encounter {
	id: string;
	name: string;
	round: number;
	activeCombatantId?: string;
	status: 'building' | 'active' | 'completed';
	combatants?: Combatant[];
	createdAt: string;
	updatedAt: string;
}

export const CONDITIONS = [
	'Отравлен',
	'Оглушён',
	'Парализован',
	'Схвачен',
	'Испуган',
	'Ослеплён',
	'Очарован',
	'Без сознания',
	'Сбит с ног',
	'Опутан',
	'Невидим',
	'Окаменевший',
	'Недееспособен',
	'Истощение'
] as const;
