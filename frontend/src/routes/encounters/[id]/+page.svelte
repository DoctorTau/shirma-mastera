<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { db } from '$lib/db';
	import type { Encounter, Combatant, Monster, CreatedCreature, PlayerCharacter } from '$lib/types';
	import { parseHpNumber } from '$lib/hp';
	import '$lib/styles/grifel.css';

	type AddMode = '' | 'monster' | 'creature' | 'player';

	let encounter = $state<Encounter | null>(null);
	let siblings = $state<Encounter[]>([]);
	let addMode = $state<AddMode>('');
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

	async function loadSiblings() {
		try {
			siblings = (await api.get<Encounter[]>('/encounters')) ?? [];
		} catch {
			siblings = await db.encounters.toArray();
		}
	}

	$effect(() => {
		if (page.params.id) load(page.params.id);
	});

	$effect(() => {
		void loadSiblings();
		void Promise.all([
			api.get<CreatedCreature[]>('/creatures').then((r) => (creatures = r)),
			api.get<PlayerCharacter[]>('/players').then((r) => (players = r))
		]);
	});

	async function createSibling() {
		const created = await api.post<Encounter>('/encounters', { name: 'Новый энкаунтер' });
		await db.encounters.put(created);
		await goto(`/encounters/${created.id}`);
	}

	async function duplicateSibling(id: string, ev: MouseEvent) {
		ev.preventDefault();
		const copy = await api.post<Encounter>(`/encounters/${id}/duplicate`);
		await db.encounters.put(copy);
		await goto(`/encounters/${copy.id}`);
	}

	async function removeSibling(id: string, ev: MouseEvent) {
		ev.preventDefault();
		if (!confirm('Удалить энкаунтер?')) return;
		await api.delete(`/encounters/${id}`);
		await db.encounters.delete(id);
		if (id === encounter?.id) {
			const next = siblings.find((s) => s.id !== id);
			if (next) await goto(`/encounters/${next.id}`);
			else await goto('/encounters');
			return;
		}
		void loadSiblings();
	}

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
		void loadSiblings();
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
		void loadSiblings();
	}

	async function startCombat() {
		if (!encounter) return;
		await api.post(`/encounters/${encounter.id}/combat/start`);
		await goto(`/combat/${encounter.id}`);
	}

	function toggleAddMode(mode: AddMode) {
		addMode = addMode === mode ? '' : mode;
	}

	// Group same-source combatants (e.g. three goblins from the same monster)
	// into one row with a "×N" badge, matching how a DM thinks about a fight.
	type Row =
		| { kind: 'group'; key: string; label: string; sub: string; meta: string; count: number; combatants: Combatant[] }
		| { kind: 'single'; combatant: Combatant };

	const rows = $derived.by((): Row[] => {
		const list = encounter?.combatants ?? [];
		const groups = new Map<string, Combatant[]>();
		const order: string[] = [];
		for (const c of list) {
			if (c.isPc || c.sourceType !== 'monster') continue;
			const key = `${c.sourceType}:${c.sourceId ?? c.displayName}`;
			if (!groups.has(key)) {
				groups.set(key, []);
				order.push(key);
			}
			groups.get(key)!.push(c);
		}
		const grouped = new Set(list.filter((c) => !c.isPc && c.sourceType === 'monster'));
		const result: Row[] = [];
		for (const key of order) {
			const members = groups.get(key)!;
			if (members.length > 1) {
				result.push({
					kind: 'group',
					key,
					label: baseName(members[0].displayName),
					sub: `${members.length} шт.`,
					meta: members.map((m) => m.displayName).join(' · '),
					count: members.length,
					combatants: members
				});
			} else {
				result.push({ kind: 'single', combatant: members[0] });
				grouped.delete(members[0]);
			}
		}
		for (const c of list) {
			if (c.isPc || c.sourceType !== 'monster') result.push({ kind: 'single', combatant: c });
		}
		return result;
	});

	function baseName(name: string): string {
		return name.replace(/\s*\d+\s*$/, '').trim() || name;
	}

	const STATUS_LABEL: Record<Encounter['status'], string> = {
		building: 'черновик',
		active: 'идёт бой',
		completed: 'завершён'
	};

	function siblingSub(s: Encounter): string {
		if (s.id === encounter?.id && encounter.combatants) {
			const n = encounter.combatants.length;
			return `${n} ${n === 1 ? 'участник' : 'участников'}`;
		}
		return STATUS_LABEL[s.status];
	}
</script>

{#if !encounter}
	<p class="gr-loading-msg">Загрузка...</p>
{:else}
	<div class="gr-encounters-screen">
		<aside class="sidebar">
			<div class="sidebar-head">
				<span class="sidebar-label">Сохранённые</span>
				<button class="new-btn" onclick={createSibling}>+ Новый</button>
			</div>
			{#each siblings as s (s.id)}
				<a
					class="saved-item"
					class:saved-item-active={s.id === encounter.id}
					href={`/encounters/${s.id}`}
				>
					<div class="saved-name">{s.name}</div>
					<div class="saved-row">
						<span class="saved-sub">{siblingSub(s)}</span>
						<span class="saved-actions">
							<button
								class="saved-action-btn"
								title="Дублировать"
								onclick={(e) => duplicateSibling(s.id, e)}>⧉</button
							>
							<button
								class="saved-action-btn"
								title="Удалить"
								onclick={(e) => removeSibling(s.id, e)}>✕</button
							>
						</span>
					</div>
				</a>
			{:else}
				<p class="sidebar-empty">Пока нет энкаунтеров</p>
			{/each}
		</aside>

		<div class="constructor">
			<div class="header">
				<input
					class="name-input"
					value={encounter.name}
					onchange={(e) => rename((e.target as HTMLInputElement).value)}
				/>
				<span class="header-tag">конструктор энкаунтера</span>
				<button
					class="gr-start-btn"
					onclick={startCombat}
					disabled={!encounter.combatants?.length}
				>
					▶ Запустить бой
				</button>
			</div>

			<div class="action-row">
				<button class="action-pill action-pill-filled" onclick={() => toggleAddMode('monster')}
					>+ Из бестиария</button
				>
				<button class="action-pill" onclick={() => toggleAddMode('creature')}>+ Свой монстр</button>
				<button class="action-pill" onclick={() => toggleAddMode('player')}>+ Игрок</button>
			</div>

			{#if addMode === 'monster'}
				<div class="add-panel">
					<input placeholder="Поиск монстра…" bind:value={monsterSearch} oninput={onMonsterSearchInput} />
					<input type="number" min="1" bind:value={monsterCount} class="count-input" />
					{#if monsterResults.length}
						<ul>
							{#each monsterResults as m (m.id)}
								<li>
									<span>{m.nameRu} <span class="dim">({m.edition}, CR {m.cr})</span></span>
									<button class="gr-add-btn" onclick={() => addMonster(m)}>Добавить</button>
								</li>
							{/each}
						</ul>
					{/if}
				</div>
			{:else if addMode === 'creature'}
				<div class="add-panel">
					{#if creatures.length}
						<ul>
							{#each creatures as c (c.id)}
								<li>
									<span>{c.nameRu}</span>
									<button class="gr-add-btn" onclick={() => addCreature(c)}>Добавить</button>
								</li>
							{/each}
						</ul>
					{:else}
						<p class="dim">Пока нет своих существ.</p>
					{/if}
				</div>
			{:else if addMode === 'player'}
				<div class="add-panel">
					{#if players.length}
						<ul>
							{#each players as p (p.id)}
								<li>
									<span>{p.name}</span>
									<button class="gr-add-btn" onclick={() => addPlayer(p)}>Добавить</button>
								</li>
							{/each}
						</ul>
					{:else}
						<p class="dim">Пока нет игроков.</p>
					{/if}
				</div>
			{/if}

			<ul class="participant-list">
				{#each rows as row (row.kind === 'group' ? row.key : row.combatant.id)}
					{#if row.kind === 'group'}
						<li class="participant participant-monster">
							<span class="badge badge-count">×{row.count}</span>
							<div class="participant-main">
								<span class="participant-name">{row.label}</span>
								<span class="participant-sub">{row.sub}</span>
							</div>
							<span class="participant-meta">{row.meta}</span>
							<button
								class="gr-remove-btn"
								onclick={() => row.combatants.forEach((c) => removeCombatant(c))}
							>
								Удалить
							</button>
						</li>
					{:else}
						{@const c = row.combatant}
						<li
							class="participant"
							class:participant-player={c.isPc}
							class:participant-custom={c.sourceType === 'created_creature'}
						>
							<span class="badge" class:badge-player={c.isPc} class:badge-custom={c.sourceType === 'created_creature'}>
								{c.isPc ? '♞' : c.sourceType === 'created_creature' ? '✎' : '⚔'}
							</span>
							<div class="participant-main">
								<span class="participant-name">{c.displayName}</span>
								<span class="participant-sub">
									{c.isPc ? 'игрок' : c.sourceType === 'created_creature' ? 'свой монстр' : 'бестиарий'}
								</span>
							</div>
							<span class="participant-meta">HP {c.maxHp ?? '—'}</span>
							<button class="gr-remove-btn" onclick={() => removeCombatant(c)}>Удалить</button>
						</li>
					{/if}
				{:else}
					<li class="empty">Пока никого нет — добавьте участников выше</li>
				{/each}
			</ul>
		</div>
	</div>
{/if}

<style>
	.gr-loading-msg {
		font-family: var(--gr-font-body);
		color: var(--gr-ink-muted);
	}

	.gr-encounters-screen {
		display: flex;
		margin: calc(var(--gr-space-xl) * -1);
		height: calc(100vh - 54px);
		font-family: var(--gr-font-body);
		color: var(--gr-ink);
	}

	.sidebar {
		width: 268px;
		flex: none;
		overflow-y: auto;
		background: var(--gr-parchment-footer);
		border-right: 1.5px solid var(--gr-parchment-border);
		padding: var(--gr-space-lg) var(--gr-space-md);
		display: flex;
		flex-direction: column;
		gap: var(--gr-space-sm);
	}
	.sidebar-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 0.25rem 0.25rem;
	}
	.sidebar-label {
		font-family: var(--gr-font-display);
		font-size: 0.625rem;
		letter-spacing: 0.16em;
		text-transform: uppercase;
		color: var(--gr-ink-faint);
	}
	.new-btn {
		font-family: var(--gr-font-display);
		font-size: 0.6875rem;
		font-weight: 600;
		color: var(--gr-accent);
		background: none;
		border: none;
		cursor: pointer;
	}
	.saved-item {
		display: block;
		border-radius: var(--gr-radius-lg);
		background: var(--gr-parchment-card);
		border: 1.5px solid var(--gr-parchment-border-strong);
		padding: 0.6875rem 0.8125rem;
		text-decoration: none;
		color: inherit;
	}
	.saved-item-active {
		background: var(--gr-accent);
		border-color: var(--gr-accent);
		color: var(--gr-cream);
	}
	.saved-name {
		font-family: var(--gr-font-display);
		font-weight: 700;
		font-size: 0.875rem;
	}
	.saved-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--gr-space-sm);
		margin-top: 0.15rem;
	}
	.saved-sub {
		font-size: 0.6875rem;
		font-style: italic;
		opacity: 0.85;
	}
	.saved-item-active .saved-sub {
		color: var(--gr-cream-soft);
	}
	.saved-item:not(.saved-item-active) .saved-sub {
		color: var(--gr-ink-muted);
	}
	.saved-actions {
		display: flex;
		gap: 0.25rem;
		flex: none;
	}
	.saved-action-btn {
		font-size: 0.6875rem;
		line-height: 1;
		width: 1.375rem;
		height: 1.375rem;
		border-radius: var(--gr-radius-sm);
		border: 1px solid rgba(255, 255, 255, 0.4);
		background: rgba(255, 255, 255, 0.12);
		color: inherit;
		cursor: pointer;
		opacity: 0.8;
	}
	.saved-item:not(.saved-item-active) .saved-action-btn {
		border-color: var(--gr-parchment-border-strong);
		background: var(--gr-parchment-panel);
	}
	.sidebar-empty {
		color: var(--gr-ink-muted);
		font-style: italic;
		font-size: 0.8125rem;
		padding: 0.5rem 0.25rem;
	}

	.constructor {
		flex: 1;
		min-width: 0;
		min-height: 0;
		padding: 1.125rem 1.375rem;
		overflow-y: auto;
	}

	.header {
		display: flex;
		align-items: baseline;
		gap: var(--gr-space-md);
		margin-bottom: var(--gr-space-xs);
		flex-wrap: wrap;
	}
	.name-input {
		font-family: var(--gr-font-display);
		font-size: 1.4rem;
		font-weight: 700;
		background: none;
		border: none;
		border-bottom: 1.5px solid transparent;
		color: var(--gr-ink-soft);
		padding: 0.15rem 0;
		min-width: 0;
	}
	.name-input:hover,
	.name-input:focus {
		border-bottom-color: var(--gr-parchment-border-strong);
	}
	.header-tag {
		font-size: 0.8125rem;
		color: var(--gr-ink-muted);
		font-style: italic;
	}
	.gr-start-btn {
		margin-left: auto;
		font-family: var(--gr-font-display);
		font-size: 0.6875rem;
		font-weight: 600;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--gr-accent);
		background: none;
		border: 1.3px solid var(--gr-accent);
		border-radius: var(--gr-radius-md);
		padding: 0.375rem 0.75rem;
		cursor: pointer;
		white-space: nowrap;
	}
	.gr-start-btn:disabled {
		opacity: 0.5;
		cursor: default;
	}

	.action-row {
		display: flex;
		gap: var(--gr-space-xs);
		margin: 0.875rem 0 1rem;
		flex-wrap: wrap;
	}
	.action-pill {
		font-family: var(--gr-font-display);
		font-size: 0.6875rem;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		border-radius: var(--gr-radius-lg);
		padding: 0.5rem 0.8125rem;
		border: 1.3px solid var(--gr-parchment-border-strong);
		background: none;
		color: var(--gr-ink-soft);
		cursor: pointer;
	}
	.action-pill-filled {
		background: var(--gr-accent);
		border-color: var(--gr-accent);
		color: var(--gr-cream);
	}

	.add-panel {
		margin-bottom: var(--gr-space-md);
		padding: 0.625rem;
		background: var(--gr-parchment-card);
		border: 1px solid var(--gr-parchment-border);
		border-radius: var(--gr-radius-md);
	}
	.add-panel input {
		font-family: var(--gr-font-body);
		padding: 0.4rem 0.5rem;
		border: 1px solid var(--gr-parchment-border-strong);
		border-radius: var(--gr-radius-sm);
		background: white;
		color: var(--gr-ink);
	}
	.count-input {
		width: 3.5rem;
		margin-left: 0.4rem;
	}
	.add-panel ul {
		list-style: none;
		padding: 0;
		margin: 0.5rem 0 0;
	}
	.add-panel li {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.3rem 0;
		font-size: 0.85rem;
	}
	.dim {
		color: var(--gr-ink-faint);
		font-size: 0.85rem;
	}

	.participant-list {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: var(--gr-space-xs);
	}
	.participant {
		display: flex;
		align-items: center;
		gap: var(--gr-space-md);
		background: var(--gr-parchment-card);
		border: 1.5px solid var(--gr-parchment-border-strong);
		border-radius: var(--gr-radius-lg);
		padding: 0.6875rem 0.875rem;
	}
	.participant-player {
		background: var(--gr-tag-player-bg);
		border-color: var(--gr-tag-player-border);
	}
	.participant-custom {
		border-style: dashed;
	}
	.badge {
		flex: none;
		width: 30px;
		height: 30px;
		border-radius: var(--gr-radius-md);
		background: var(--gr-ink-muted);
		color: var(--gr-cream);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.8125rem;
		font-weight: 700;
	}
	.badge-count {
		background: var(--gr-accent);
	}
	.badge-player {
		background: var(--gr-player-bg);
		color: var(--gr-player-fg);
	}
	.badge-custom {
		background: var(--gr-cream-dim);
		color: var(--gr-ink-soft);
	}
	.participant-main {
		flex: 1;
		min-width: 0;
		display: flex;
		align-items: baseline;
		gap: 0.5rem;
	}
	.participant-name {
		font-family: var(--gr-font-display);
		font-weight: 700;
		font-size: 0.9375rem;
		color: var(--gr-ink);
	}
	.participant-sub {
		font-size: 0.75rem;
		font-style: italic;
		color: var(--gr-ink-faint);
	}
	.participant-meta {
		font-size: 0.75rem;
		color: var(--gr-ink-muted);
		white-space: nowrap;
	}
	.gr-remove-btn {
		flex: none;
		font-family: var(--gr-font-display);
		font-size: 0.6875rem;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		background: none;
		border: 1.3px solid var(--gr-parchment-border-strong);
		color: var(--gr-ink-muted);
		border-radius: var(--gr-radius-sm);
		padding: 0.3rem 0.6rem;
		cursor: pointer;
	}
	.empty {
		color: var(--gr-ink-muted);
		font-style: italic;
		padding: 0.5rem 0.25rem;
	}
</style>
