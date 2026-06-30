<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { db } from '$lib/db';
	import type { Encounter, Combatant, Monster, CreatedCreature, PlayerCharacter } from '$lib/types';
	import { parseHpNumber } from '$lib/hp';
	import '$lib/styles/grifel.css';

	let encounter = $state<Encounter | null>(null);
	let monsterSearch = $state('');
	let monsterResults = $state<Monster[]>([]);
	let creatures = $state<CreatedCreature[]>([]);
	let players = $state<PlayerCharacter[]>([]);
	let monsterCount = $state(1);
	let searchTimer: ReturnType<typeof setTimeout>;

	async function load(id: string) {
		encounter = await api.get<Encounter>(`/encounters/${id}`);
		void db.encounters.put(encounter);
	}

	$effect(() => {
		if (page.params.id) load(page.params.id);
	});

	$effect(() => {
		void Promise.all([
			api.get<CreatedCreature[]>('/creatures').then((r) => (creatures = r)),
			api.get<PlayerCharacter[]>('/players').then((r) => (players = r))
		]);
	});

	function onMonsterSearchInput() {
		clearTimeout(searchTimer);
		searchTimer = setTimeout(async () => {
			if (!monsterSearch) {
				monsterResults = [];
				return;
			}
			monsterResults = await api.get<Monster[]>(
				`/monsters?search=${encodeURIComponent(monsterSearch)}&limit=20`
			);
		}, 300);
	}

	async function refreshEncounter() {
		if (!encounter) return;
		encounter = await api.get<Encounter>(`/encounters/${encounter.id}`);
		void db.encounters.put(encounter);
	}

	async function addMonster(m: Monster) {
		if (!encounter) return;
		await api.post(`/encounters/${encounter.id}/combatants`, {
			sourceType: 'monster',
			sourceId: m.id,
			monsterEdition: m.edition,
			displayName: m.nameRu,
			maxHp: parseHpNumber(m.statblock.hitPoints),
			count: monsterCount
		});
		monsterCount = 1;
		await refreshEncounter();
	}

	async function addCreature(c: CreatedCreature) {
		if (!encounter) return;
		await api.post(`/encounters/${encounter.id}/combatants`, {
			sourceType: 'created_creature',
			sourceId: c.id,
			displayName: c.nameRu,
			maxHp: parseHpNumber(c.statblock.hitPoints),
			count: 1
		});
		await refreshEncounter();
	}

	async function addPlayer(p: PlayerCharacter) {
		if (!encounter) return;
		await api.post(`/encounters/${encounter.id}/combatants`, {
			sourceType: 'player_character',
			sourceId: p.id,
			displayName: p.name,
			maxHp: p.maxHp,
			isPc: true,
			count: 1
		});
		await refreshEncounter();
	}

	async function removeCombatant(c: Combatant) {
		if (!encounter) return;
		await api.delete(`/encounters/${encounter.id}/combatants/${c.id}`);
		await refreshEncounter();
	}

	async function rename(name: string) {
		if (!encounter) return;
		encounter = await api.put<Encounter>(`/encounters/${encounter.id}`, { name });
	}

	async function startCombat() {
		if (!encounter) return;
		await api.post(`/encounters/${encounter.id}/combat/start`);
		await goto(`/combat/${encounter.id}`);
	}
</script>

{#if !encounter}
	<p class="gr-loading-msg">Загрузка...</p>
{:else}
	<div class="gr-encounter-editor">
		<div class="header">
			<input
				class="name-input"
				value={encounter.name}
				onchange={(e) => rename((e.target as HTMLInputElement).value)}
			/>
			<button class="gr-start-btn" onclick={startCombat} disabled={!encounter.combatants?.length}>
				▶ Запустить бой
			</button>
		</div>

		<section class="participants">
			<h2>Участники</h2>
			<ul class="combatant-list">
				{#each encounter.combatants ?? [] as c (c.id)}
					<li>
						<span class="name">{c.displayName}</span>
						<span class="hp">HP {c.maxHp ?? '—'}</span>
						<span class="tag">{c.sourceType}</span>
						<button class="gr-remove-btn" onclick={() => removeCombatant(c)}>Удалить</button>
					</li>
				{:else}
					<li class="empty">Пока никого нет</li>
				{/each}
			</ul>
		</section>

		<section class="add-section">
			<h2>+ Из бестиария</h2>
			<div class="row">
				<input placeholder="Поиск монстра" bind:value={monsterSearch} oninput={onMonsterSearchInput} />
				<input type="number" min="1" bind:value={monsterCount} class="count-input" />
			</div>
			<ul class="pick-list">
				{#each monsterResults as m (m.id)}
					<li>
						<span>{m.nameRu} <span class="dim">({m.edition}, CR {m.cr})</span></span>
						<button class="gr-add-btn" onclick={() => addMonster(m)}>Добавить</button>
					</li>
				{/each}
			</ul>
		</section>

		<section class="add-section">
			<h2>+ Своё существо</h2>
			<ul class="pick-list">
				{#each creatures as c (c.id)}
					<li>
						<span>{c.nameRu}</span>
						<button class="gr-add-btn" onclick={() => addCreature(c)}>Добавить</button>
					</li>
				{/each}
			</ul>
		</section>

		<section class="add-section">
			<h2>+ Игрок</h2>
			<ul class="pick-list">
				{#each players as p (p.id)}
					<li>
						<span>{p.name}</span>
						<button class="gr-add-btn" onclick={() => addPlayer(p)}>Добавить</button>
					</li>
				{/each}
			</ul>
		</section>
	</div>
{/if}

<style>
	.gr-loading-msg {
		font-family: var(--gr-font-body);
		color: var(--gr-ink-muted);
	}
	.gr-encounter-editor {
		font-family: var(--gr-font-body);
		color: var(--gr-ink);
		max-width: 44rem;
	}
	.header {
		display: flex;
		gap: var(--gr-space-md);
		align-items: center;
		margin-bottom: var(--gr-space-lg);
	}
	.name-input {
		flex: 1;
		font-family: var(--gr-font-display);
		font-size: 1.3rem;
		font-weight: 700;
		padding: 0.5rem 0.7rem;
		background: var(--gr-input-bg);
		border: 1.5px solid var(--gr-input-border);
		color: var(--gr-ink);
		border-radius: var(--gr-radius-md);
	}
	.gr-start-btn {
		font-family: var(--gr-font-display);
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		background: var(--gr-accent);
		box-shadow: inset 0 -2px 0 var(--gr-accent-shadow);
		color: var(--gr-cream);
		border: none;
		padding: 0.65rem 1.2rem;
		border-radius: var(--gr-radius-md);
		cursor: pointer;
		white-space: nowrap;
	}
	.gr-start-btn:disabled {
		opacity: 0.5;
		cursor: default;
	}
	section {
		margin-bottom: var(--gr-space-xl);
	}
	h2 {
		font-family: var(--gr-font-display);
		font-size: 0.8rem;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--gr-accent);
		border-bottom: 1.5px solid var(--gr-accent);
		padding-bottom: 0.3rem;
		margin: 0 0 var(--gr-space-sm);
	}
	.combatant-list,
	.pick-list {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: var(--gr-space-xs);
	}
	.combatant-list li,
	.pick-list li {
		display: flex;
		gap: var(--gr-space-md);
		align-items: center;
		padding: 0.55rem 0.8rem;
		border-radius: var(--gr-radius-md);
		background: var(--gr-parchment-card);
		border: 1.5px solid var(--gr-parchment-border-strong);
	}
	.combatant-list li .name {
		font-family: var(--gr-font-display);
		font-weight: 700;
	}
	.combatant-list li .hp,
	.combatant-list li .tag {
		color: var(--gr-ink-muted);
		font-size: 0.85rem;
	}
	.combatant-list li button,
	.pick-list li button {
		margin-left: auto;
	}
	.gr-remove-btn,
	.gr-add-btn {
		font-family: var(--gr-font-display);
		font-size: 0.6875rem;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		background: none;
		border: 1.3px solid var(--gr-parchment-border-strong);
		color: var(--gr-ink-muted);
		border-radius: var(--gr-radius-sm);
		padding: 0.35rem 0.65rem;
		cursor: pointer;
	}
	.gr-remove-btn {
		border-color: var(--gr-accent);
		color: var(--gr-accent);
	}
	.gr-add-btn {
		border-color: var(--gr-accent);
		color: var(--gr-cream);
		background: var(--gr-accent);
	}
	.empty {
		color: var(--gr-ink-muted);
		font-style: italic;
		padding: 0.5rem 0.25rem;
	}
	.row {
		display: flex;
		gap: var(--gr-space-sm);
		margin-bottom: var(--gr-space-sm);
	}
	.row input {
		font-family: var(--gr-font-body);
		padding: 0.55rem 0.7rem;
		background: var(--gr-input-bg);
		border: 1.5px solid var(--gr-input-border);
		color: var(--gr-ink);
		border-radius: var(--gr-radius-md);
	}
	.row input:first-child {
		flex: 1;
	}
	.count-input {
		width: 4rem;
	}
	.dim {
		color: var(--gr-ink-faint);
		font-size: 0.85rem;
	}
</style>
