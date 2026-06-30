import Dexie, { type Table } from 'dexie';
import type { Monster, CreatedCreature, PlayerCharacter, Encounter, Combatant } from './types';

export interface MutationQueueItem {
	id?: number;
	method: 'POST' | 'PUT' | 'PATCH' | 'DELETE';
	path: string;
	body?: unknown;
	createdAt: string;
}

export interface CatalogEntryRow {
	dndsuId: number;
	edition: string;
	slug: string;
	nameRu: string;
	nameEn: string;
	cr: string;
	type: string;
	url: string;
	isUniqueNpc: boolean;
}

class DmCompanionDB extends Dexie {
	monsters!: Table<Monster, string>;
	catalogEntries!: Table<CatalogEntryRow, [number, string]>;
	createdCreatures!: Table<CreatedCreature, string>;
	playerCharacters!: Table<PlayerCharacter, string>;
	encounters!: Table<Encounter, string>;
	combatants!: Table<Combatant, string>;
	mutationQueue!: Table<MutationQueueItem, number>;

	constructor() {
		super('dm-companion');
		this.version(1).stores({
			monsters: 'id, slug, edition, nameRu, nameEn, type, cr',
			catalogEntries: '[dndsuId+edition], nameRu, nameEn, type, cr',
			createdCreatures: 'id, nameRu, nameEn, updatedAt',
			playerCharacters: 'id, name, updatedAt',
			encounters: 'id, name, status, updatedAt',
			combatants: 'id, encounterId, sortOrder',
			mutationQueue: '++id, createdAt'
		});
	}
}

export const db = new DmCompanionDB();

export async function requestPersistentStorage() {
	if (navigator.storage?.persist) {
		await navigator.storage.persist();
	}
}

/** Queues a write for background sync and lets the caller apply it locally first (optimistic). */
export async function enqueueMutation(item: Omit<MutationQueueItem, 'id' | 'createdAt'>) {
	await db.mutationQueue.add({ ...item, createdAt: new Date().toISOString() });
}
