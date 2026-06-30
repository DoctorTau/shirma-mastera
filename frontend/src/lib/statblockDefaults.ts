import type { StatBlock } from './types';

export function emptyAbilities() {
	const a = { score: 10, mod: 0 };
	return { str: { ...a }, dex: { ...a }, con: { ...a }, int: { ...a }, wis: { ...a }, cha: { ...a } };
}

export function emptyStatBlock(): StatBlock {
	return { abilities: emptyAbilities() };
}
