<script lang="ts">
	import '$lib/styles/grifel.css';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api, ApiError } from '$lib/api';
	import { db, enqueueMutation } from '$lib/db';
	import { flushMutationQueue } from '$lib/sync';
	import type { Encounter, Combatant, Monster, CreatedCreature, StatBlock } from '$lib/types';
	import { CONDITIONS } from '$lib/types';
	import { parseHpNumber } from '$lib/hp';
	import GrifelStatBlock from '$lib/components/GrifelStatBlock.svelte';
	import { slide, scale, fade } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	let encounter = $state<Encounter | null>(null);
	let selectedId = $state<string | null>(null);
	let selectedStatblock = $state<StatBlock | null>(null);
	let selectedLoading = $state(false);
	let selectedSourceUrl = $state<string | undefined>(undefined);
	let showAddPanel = $state(false);
	let addSearch = $state('');
	let addResults = $state<Monster[]>([]);
	let addCount = $state(1);
	let searchTimer: ReturnType<typeof setTimeout>;
	let selectedNameEn = $state<string | undefined>(undefined);
	let selectedImageUrl = $state<string | undefined>(undefined);
	let acByCombatant = $state<Record<string, string>>({});
	let dmgAmount = $state(5);
	const statblockCache = new Map<string, StatBlock & { nameEn?: string; imageUrl?: string }>();
	const playerCache = new Map<string, { ac?: number }>();

	const sorted = $derived(
		[...(encounter?.combatants ?? [])].sort((a, b) => {
			const ai = a.initiative ?? -Infinity;
			const bi = b.initiative ?? -Infinity;
			if (bi !== ai) return bi - ai;
			return a.sortOrder - b.sortOrder;
		})
	);

	const selected = $derived(
		sorted.find((c) => c.id === selectedId) ??
			sorted.find((c) => c.id === encounter?.activeCombatantId) ??
			sorted[0] ??
			null
	);

	async function load(id: string) {
		try {
			encounter = await api.get<Encounter>(`/encounters/${id}`);
			void db.encounters.put(encounter);
			if (encounter.combatants) void db.combatants.bulkPut(encounter.combatants);
		} catch {
			const cached = await db.encounters.get(id);
			if (cached) {
				cached.combatants = await db.combatants.where('encounterId').equals(id).toArray();
				encounter = cached;
			}
		}
	}

	$effect(() => {
		if (page.params.id) load(page.params.id);
	});

	async function patchCombatant(c: Combatant, body: Partial<Combatant>) {
		if (!encounter) return;
		Object.assign(c, body);
		encounter.combatants = encounter.combatants?.map((x) => (x.id === c.id ? c : x));
		void db.combatants.put(c);
		const path = `/encounters/${encounter.id}/combatants/${c.id}`;
		try {
			await api.patch(path, body);
		} catch (err) {
			if (!(err instanceof ApiError)) {
				await enqueueMutation({ method: 'PATCH', path, body });
			}
		}
	}

	function adjustHp(c: Combatant, delta: number) {
		const current = c.currentHp ?? c.maxHp ?? 0;
		let next = current + delta;
		if (c.maxHp != null) next = Math.min(next, c.maxHp);
		void patchCombatant(c, { currentHp: next });
	}

	function adjustTempHp(c: Combatant, delta: number) {
		const next = Math.max(0, (c.tempHp ?? 0) + delta);
		void patchCombatant(c, { tempHp: next });
	}

	function toggleCondition(c: Combatant, condition: string) {
		const has = c.conditions.includes(condition);
		const next = has ? c.conditions.filter((x) => x !== condition) : [...c.conditions, condition];
		void patchCombatant(c, { conditions: next });
	}

	function setInitiative(c: Combatant, value: number) {
		void patchCombatant(c, { initiative: value });
	}

	async function fetchStatblock(
		c: Combatant
	): Promise<{ statblock: (StatBlock & { nameEn?: string; imageUrl?: string }) | null; sourceUrl?: string }> {
		if (!c.sourceId) return { statblock: null };
		const cacheKey = c.sourceId + (c.monsterEdition ?? '');
		if (statblockCache.has(cacheKey)) {
			return { statblock: statblockCache.get(cacheKey)!, sourceUrl: undefined };
		}
		try {
			if (c.sourceType === 'monster') {
				const m = await api.get<Monster>(`/monsters/${c.sourceId}`);
				const sb = { ...m.statblock, nameEn: m.nameEn, imageUrl: m.imageUrl };
				statblockCache.set(cacheKey, sb);
				return { statblock: sb, sourceUrl: m.sourceUrl };
			}
			if (c.sourceType === 'created_creature') {
				const cc = await api.get<CreatedCreature>(`/creatures/${c.sourceId}`);
				const sb = { ...cc.statblock, nameEn: cc.nameEn };
				statblockCache.set(cacheKey, sb);
				return { statblock: sb };
			}
		} catch {
			return { statblock: null };
		}
		return { statblock: null };
	}

	async function loadAc(c: Combatant) {
		if (acByCombatant[c.id] !== undefined) return;
		if (c.sourceType === 'player_character' && c.sourceId) {
			try {
				let p = playerCache.get(c.sourceId);
				if (!p) {
					p = await api.get<{ ac?: number }>(`/players/${c.sourceId}`);
					playerCache.set(c.sourceId, p);
				}
				if (p.ac != null) acByCombatant = { ...acByCombatant, [c.id]: String(p.ac) };
			} catch {
				/* no AC available offline */
			}
			return;
		}
		if (c.sourceType === 'monster' || c.sourceType === 'created_creature') {
			const { statblock } = await fetchStatblock(c);
			if (statblock?.armorClass) acByCombatant = { ...acByCombatant, [c.id]: statblock.armorClass };
		}
	}

	$effect(() => {
		for (const c of sorted) void loadAc(c);
	});

	$effect(() => {
		const c = selected;
		if (!c) {
			selectedStatblock = null;
			selectedSourceUrl = undefined;
			selectedNameEn = undefined;
			selectedImageUrl = undefined;
			return;
		}
		selectedLoading = true;
		fetchStatblock(c).then(({ statblock, sourceUrl }) => {
			if (selected?.id === c.id) {
				selectedStatblock = statblock;
				selectedSourceUrl = sourceUrl;
				selectedNameEn = statblock?.nameEn;
				selectedImageUrl = statblock?.imageUrl;
				selectedLoading = false;
			}
		});
	});

	async function patchState(body: { round?: number; activeCombatantId?: string }) {
		if (!encounter) return;
		Object.assign(encounter, body);
		void db.encounters.put(encounter);
		try {
			await api.patch(`/encounters/${encounter.id}/combat/state`, body);
		} catch (err) {
			if (!(err instanceof ApiError)) {
				await enqueueMutation({
					method: 'PATCH',
					path: `/encounters/${encounter.id}/combat/state`,
					body
				});
			}
		}
	}

	function nextTurn() {
		if (!encounter || sorted.length === 0) return;
		const idx = sorted.findIndex((c) => c.id === encounter!.activeCombatantId);
		const nextIdx = idx + 1;
		selectedId = null;
		if (nextIdx >= sorted.length) {
			void patchState({ round: encounter.round + 1, activeCombatantId: sorted[0].id });
		} else {
			void patchState({ activeCombatantId: sorted[nextIdx].id });
		}
	}

	function prevTurn() {
		if (!encounter || sorted.length === 0) return;
		const idx = sorted.findIndex((c) => c.id === encounter!.activeCombatantId);
		const prevIdx = idx - 1;
		selectedId = null;
		if (prevIdx < 0) {
			void patchState({
				round: Math.max(1, encounter.round - 1),
				activeCombatantId: sorted[sorted.length - 1].id
			});
		} else {
			void patchState({ activeCombatantId: sorted[prevIdx].id });
		}
	}

	async function endCombat() {
		if (!encounter || !confirm('Завершить бой?')) return;
		await api.post(`/encounters/${encounter.id}/combat/end`);
		await goto('/encounters');
	}

	function onAddSearchInput() {
		clearTimeout(searchTimer);
		searchTimer = setTimeout(async () => {
			if (!addSearch) {
				addResults = [];
				return;
			}
			addResults = await api.get<Monster[]>(`/monsters?search=${encodeURIComponent(addSearch)}&limit=15`);
		}, 300);
	}

	async function addMonsterOnTheFly(m: Monster) {
		if (!encounter) return;
		const created = await api.post<Combatant[]>(`/encounters/${encounter.id}/combatants`, {
			sourceType: 'monster',
			sourceId: m.id,
			monsterEdition: m.edition,
			displayName: m.nameRu,
			maxHp: parseHpNumber(m.statblock.hitPoints),
			count: addCount
		});
		encounter.combatants = [...(encounter.combatants ?? []), ...created];
		void db.combatants.bulkPut(created);
		addCount = 1;
		addSearch = '';
		addResults = [];
		showAddPanel = false;
	}

	if (typeof window !== 'undefined') {
		window.addEventListener('online', () => void flushMutationQueue());
	}

	function hpFraction(c: Combatant): number {
		if (!c.maxHp) return 1;
		return Math.max(0, Math.min(1, (c.currentHp ?? c.maxHp) / c.maxHp));
	}

	function hpTier(c: Combatant): 'good' | 'mid' | 'low' {
		const frac = hpFraction(c);
		if (frac > 0.5) return 'good';
		if (frac > 0.25) return 'mid';
		return 'low';
	}

	function isDown(c: Combatant): boolean {
		return (c.currentHp ?? 1) <= 0;
	}
</script>

{#if !encounter}
	<div class="gr-loading">Загрузка...</div>
{:else}
	<div class="gr-app">
		<div class="gr-combat-toolbar">
			<div class="gr-round">
				<span class="gr-round-label">Раунд</span>
				<span class="gr-round-num">{encounter.round}</span>
			</div>
			<button class="gr-icon-btn" title="К энкаунтерам" onclick={() => goto('/encounters')}>⤺</button>
			<button class="gr-icon-btn" title="Завершить бой" onclick={endCombat}>⚙</button>
		</div>

		<div class="gr-layout">
			<div class="gr-left">
				<div class="gr-left-head">
					<div>
						<h1>{encounter.name}</h1>
						<p class="gr-sub">{sorted.length} участников · порядок инициативы</p>
					</div>
					<button class="gr-outline-btn" onclick={() => (showAddPanel = !showAddPanel)}>+ Участник</button>
				</div>

				{#if showAddPanel}
					<div class="gr-add-panel" transition:slide={{ duration: 180, easing: cubicOut }}>
						<input placeholder="Поиск монстра…" bind:value={addSearch} oninput={onAddSearchInput} />
						<input type="number" min="1" bind:value={addCount} class="gr-add-count" />
						{#if addResults.length}
							<ul>
								{#each addResults as m (m.id)}
									<li>
										<span>{m.nameRu} ({m.edition})</span>
										<button onclick={() => addMonsterOnTheFly(m)}>Добавить</button>
									</li>
								{/each}
							</ul>
						{/if}
					</div>
				{/if}

				<ul class="gr-init-list">
					{#each sorted as c (c.id)}
						{@const active = c.id === encounter.activeCombatantId}
						{@const down = isDown(c)}
						<li
							class="gr-row"
							class:active
							class:down
							class:selected={selected?.id === c.id}
						>
							<div
								class="gr-init-circle"
								class:pc={c.isPc}
								class:down
								onclick={() => (selectedId = c.id)}
								onkeydown={(e) => {
									if (e.key === 'Enter' || e.key === ' ') selectedId = c.id;
								}}
								role="button"
								tabindex="0"
								title="Показать статблок"
							>
								<input
									class="gr-init-num-input"
									type="number"
									value={c.initiative ?? ''}
									onclick={(e) => e.stopPropagation()}
									onfocus={() => (selectedId = c.id)}
									onchange={(e) => setInitiative(c, Number((e.target as HTMLInputElement).value))}
								/>
								<span class="gr-init-label">иниц</span>
							</div>

							<div class="gr-row-main">
								<div class="gr-row-top">
									<button class="gr-name-btn" onclick={() => (selectedId = c.id)}>{c.displayName}</button>
									{#if c.isPc}<span class="gr-tag gr-tag-player">Игрок</span>{/if}
									{#if active}<span class="gr-badge-hod">Ход</span>{/if}
									{#if down}<span class="gr-badge-down">Без сознания</span>{/if}
									{#if acByCombatant[c.id]}
										<span class="gr-ac">КД <b>{acByCombatant[c.id]}</b></span>
									{/if}
								</div>

								<div class="gr-hp-row">
									<div class="gr-hp-bar">
										<div class="gr-hp-fill gr-hp-{hpTier(c)}" style={`width:${hpFraction(c) * 100}%`}></div>
									</div>
									<span class="gr-hp-num"
										>{c.currentHp ?? '—'}<span class="gr-hp-max">/{c.maxHp ?? '—'}{#if c.tempHp}+{c.tempHp}{/if}</span
										></span
									>
								</div>

								{#if c.conditions.length || selected?.id === c.id}
									<div class="gr-cond-hp-row">
										{#if selected?.id === c.id}
											<div class="gr-dmg-stepper">
												<button class="gr-dmg-btn gr-dmg-minus" onclick={() => adjustHp(c, -dmgAmount)}>−</button>
												<input type="number" class="gr-dmg-amount" bind:value={dmgAmount} min="0" />
												<button class="gr-dmg-btn gr-dmg-plus" onclick={() => adjustHp(c, dmgAmount)}>+</button>
												<button
													class="gr-temp-btn"
													title="Временные хиты"
													onclick={() => adjustTempHp(c, dmgAmount)}
												>
													врем
												</button>
											</div>
										{/if}

										<div class="gr-condition-tags">
											{#each c.conditions as cond (cond)}
												<button
												class="gr-tag gr-tag-warn"
												onclick={() => toggleCondition(c, cond)}
												out:scale={{ duration: 140, start: 0.9 }}>{cond} ×</button
											>
											{/each}
											{#if selected?.id === c.id}
												<select
													class="gr-cond-select"
													onchange={(e) => {
														const v = (e.target as HTMLSelectElement).value;
														if (v) toggleCondition(c, v);
														(e.target as HTMLSelectElement).value = '';
													}}
												>
													<option value="">+ состояние</option>
													{#each CONDITIONS.filter((cond) => !c.conditions.includes(cond)) as cond (cond)}
														<option value={cond}>{cond}</option>
													{/each}
												</select>
											{/if}
										</div>
									</div>
								{/if}
							</div>
						</li>
					{:else}
						<li class="gr-empty">Нет участников</li>
					{/each}
				</ul>

				<div class="gr-turn-controls">
					<button class="gr-prev-btn" onclick={prevTurn}>◀ Пред</button>
					<button class="gr-next-btn" onclick={nextTurn}>Следующий ход ▶</button>
				</div>
			</div>

			<div class="gr-right">
				{#key selected?.id}
					<div class="gr-right-inner" in:fade={{ duration: 150 }}>
						{#if selectedLoading}
							<p class="gr-panel-msg">Загрузка...</p>
						{:else if selected && selectedStatblock}
							<GrifelStatBlock
								statblock={selectedStatblock}
								nameRu={selected.displayName}
								nameEn={selectedNameEn}
								imageUrl={selectedImageUrl}
								sourceUrl={selectedSourceUrl}
								notes={selected.notes}
								onNotesChange={(value) => selected && patchCombatant(selected, { notes: value })}
							/>
						{:else}
							<p class="gr-panel-msg">Нет статблока для этого участника.</p>
						{/if}
					</div>
				{/key}
			</div>
		</div>
	</div>
{/if}

<style>
	.gr-loading {
		padding: 2rem;
		font-family: var(--gr-font-body);
		color: var(--gr-ink);
	}

	.gr-app {
		display: flex;
		flex-direction: column;
		margin: calc(var(--gr-space-xl) * -1);
		height: calc(100vh - 54px);
		background: var(--gr-parchment-panel);
		font-family: var(--gr-font-body);
		color: var(--gr-ink);
	}

	.gr-combat-toolbar {
		flex: 0 0 auto;
		height: 44px;
		display: flex;
		align-items: center;
		gap: 18px;
		padding: 0 1.1rem;
		background: var(--gr-parchment-list);
		border-bottom: 1.5px solid var(--gr-parchment-border);
	}
	.gr-header-spacer {
		flex: 1;
	}
	.gr-round {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		line-height: 1.1;
	}
	.gr-round-label {
		font-family: var(--gr-font-display);
		font-size: 0.625rem;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: var(--gr-ink-muted);
	}
	.gr-round-num {
		font-family: var(--gr-font-display);
		font-weight: 700;
		font-size: 1.25rem;
		color: var(--gr-ink);
	}
	.gr-icon-btn {
		background: none;
		border: none;
		color: var(--gr-ink-muted);
		font-size: 1.1rem;
		cursor: pointer;
		opacity: 0.85;
		line-height: 1;
		padding: 0.25rem;
	}
	.gr-icon-btn:hover {
		opacity: 1;
		color: var(--gr-ink);
	}

	.gr-layout {
		flex: 1;
		display: flex;
		min-height: 0;
	}

	.gr-left {
		width: 41.7%;
		min-width: 22rem;
		max-width: 32rem;
		display: flex;
		flex-direction: column;
		background: var(--gr-parchment-list);
		border-right: 1.5px solid var(--gr-parchment-border);
		min-height: 0;
	}
	.gr-left-head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 0.75rem;
		padding: 1rem 1.1rem 0.75rem;
	}
	.gr-left-head h1 {
		margin: 0;
		font-family: var(--gr-font-display);
		font-size: 1.05rem;
		color: var(--gr-ink);
	}
	.gr-sub {
		margin: 0.2rem 0 0;
		font-size: 0.75rem;
		color: var(--gr-ink-muted);
	}

	.gr-outline-btn {
		font-family: var(--gr-font-display);
		font-size: 0.6875rem;
		font-weight: 600;
		letter-spacing: 0.055em;
		text-transform: uppercase;
		color: var(--gr-accent);
		background: none;
		border: 1px solid var(--gr-accent);
		border-radius: var(--gr-radius-md);
		padding: 0.375rem 0.7rem;
		cursor: pointer;
		white-space: nowrap;
	}

	.gr-add-panel {
		margin: 0 1.1rem 0.75rem;
		padding: 0.6rem;
		background: var(--gr-parchment-card);
		border: 1px solid var(--gr-parchment-border);
		border-radius: var(--gr-radius-md);
	}
	.gr-add-panel input {
		font-family: var(--gr-font-body);
		padding: 0.4rem 0.5rem;
		border: 1px solid var(--gr-parchment-border-strong);
		border-radius: var(--gr-radius-sm);
		background: white;
		color: var(--gr-ink);
	}
	.gr-add-count {
		width: 3.5rem;
		margin-left: 0.4rem;
	}
	.gr-add-panel ul {
		list-style: none;
		padding: 0;
		margin: 0.5rem 0 0;
	}
	.gr-add-panel li {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.25rem 0;
		font-size: 0.85rem;
	}

	.gr-init-list {
		flex: 1;
		overflow-y: auto;
		list-style: none;
		margin: 0;
		padding: 0 0.75rem;
	}
	.gr-row {
		position: relative;
		display: flex;
		gap: 0.7rem;
		padding: 0.7rem 0.6rem 0.7rem 0.85rem;
		margin-bottom: 0.4rem;
		border-radius: var(--gr-radius-lg);
		background: var(--gr-parchment-card);
		border: 1.5px solid var(--gr-parchment-border-strong);
		transition:
			background-color var(--gr-duration-base) ease,
			border-color var(--gr-duration-base) ease,
			box-shadow var(--gr-duration-base) ease;
	}
	.gr-row.selected:not(.active) {
		border-color: var(--gr-accent);
	}
	.gr-row.active {
		background: var(--gr-parchment-active);
		border: 2px solid var(--gr-accent);
		box-shadow: 0 4px 14px -4px rgba(122, 31, 31, 0.45);
		padding: calc(0.7rem - 0.5px) calc(0.6rem - 0.5px) calc(0.7rem - 0.5px) calc(0.85rem - 0.5px);
	}
	.gr-row.down {
		opacity: 0.65;
	}

	.gr-init-circle {
		flex: 0 0 auto;
		width: 42px;
		height: 42px;
		border-radius: var(--gr-radius-lg);
		background: var(--gr-accent);
		color: var(--gr-cream);
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		line-height: 1;
		cursor: pointer;
	}
	.gr-init-circle.pc {
		background: var(--gr-player-bg);
		color: var(--gr-player-fg);
	}
	.gr-init-circle.down {
		background: var(--gr-down-bg);
		color: var(--gr-down-fg);
	}
	.gr-init-num-input {
		width: 100%;
		font-family: var(--gr-font-display);
		font-weight: 700;
		font-size: 1.1875rem;
		text-align: center;
		background: none;
		border: none;
		color: inherit;
		padding: 0;
		-moz-appearance: textfield;
		appearance: textfield;
	}
	.gr-init-num-input::-webkit-inner-spin-button,
	.gr-init-num-input::-webkit-outer-spin-button {
		-webkit-appearance: none;
		margin: 0;
	}
	.gr-init-num-input:focus {
		outline: none;
	}
	.gr-init-label {
		font-family: var(--gr-font-display);
		font-size: 0.5rem;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		opacity: 0.85;
	}

	.gr-row-main {
		flex: 1;
		min-width: 0;
	}
	.gr-row-top {
		display: flex;
		align-items: center;
		gap: 0.4rem;
		flex-wrap: wrap;
	}
	.gr-name-btn {
		font-family: var(--gr-font-display);
		font-weight: 700;
		font-size: 1rem;
		color: var(--gr-ink);
		background: none;
		border: none;
		cursor: pointer;
		padding: 0;
	}
	.gr-ac {
		margin-left: auto;
		font-size: 0.75rem;
		color: var(--gr-ink-muted);
		white-space: nowrap;
	}
	.gr-ac b {
		color: var(--gr-ink);
	}

	.gr-tag {
		display: inline-block;
		vertical-align: middle;
		margin: 0 0.3rem 0.3rem 0;
		font-family: var(--gr-font-body);
		font-size: 0.6875rem;
		font-weight: 600;
		padding: 0.0625rem 0.4375rem;
		border-radius: var(--gr-radius-sm);
		border: 1px solid;
		cursor: default;
	}
	.gr-tag-player {
		background: var(--gr-tag-player-bg);
		border-color: var(--gr-tag-player-border);
		color: var(--gr-tag-player-fg);
	}
	.gr-tag-warn {
		background: var(--gr-tag-warn-bg);
		border-color: var(--gr-tag-warn-border);
		color: var(--gr-tag-warn-fg);
		font-weight: 400;
		cursor: pointer;
	}
	.gr-badge-hod {
		font-family: var(--gr-font-display);
		font-size: 0.5625rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		background: var(--gr-accent);
		color: var(--gr-cream);
		border-radius: 20px;
		padding: 0.1875rem 0.5rem;
	}
	.gr-badge-down {
		font-family: var(--gr-font-display);
		font-size: 0.5625rem;
		letter-spacing: 0.07em;
		text-transform: uppercase;
		background: var(--gr-ink-soft);
		color: var(--gr-cream-soft);
		border-radius: 20px;
		padding: 0.1875rem 0.5rem;
	}

	.gr-hp-row {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-top: 0.4rem;
	}
	.gr-hp-bar {
		flex: 1;
		min-width: 6rem;
		height: 11px;
		background: rgba(44, 34, 24, 0.12);
		border-radius: 6px;
		overflow: hidden;
	}
	.gr-hp-fill {
		height: 100%;
		transition:
			width 220ms var(--gr-ease-in-out),
			background-color 220ms var(--gr-ease);
	}
	.gr-hp-good {
		background: linear-gradient(var(--gr-hp-good-1), var(--gr-hp-good-2));
	}
	.gr-hp-mid {
		background: linear-gradient(var(--gr-hp-mid-1), var(--gr-hp-mid-2));
	}
	.gr-hp-low {
		background: var(--gr-hp-low);
	}
	.gr-hp-num {
		flex: 0 0 auto;
		font-size: 0.8125rem;
		font-weight: 700;
		color: var(--gr-ink-soft);
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
	}
	.gr-hp-max {
		font-weight: 400;
		color: var(--gr-ink-faint);
	}
	.gr-conditions {
		display: flex;
		gap: 0.3rem;
		flex-wrap: wrap;
		align-items: center;
		margin-top: 0.35rem;
	}
	.gr-cond-hp-row {
		margin-top: 0.35rem;
		overflow: hidden;
	}
	.gr-condition-tags {
		display: block;
	}
	.gr-cond-select {
		display: inline-block;
		vertical-align: middle;
		margin: 0 0.3rem 0.3rem 0;
		font-family: var(--gr-font-body);
		font-size: 0.6875rem;
		font-weight: 600;
		padding: 0.0625rem 0.4375rem;
		border-radius: var(--gr-radius-sm);
		border: 1px solid var(--gr-parchment-border-strong);
		background: var(--gr-parchment-card);
		color: var(--gr-ink-muted);
	}

	.gr-dmg-stepper {
		float: right;
		display: flex;
		align-items: stretch;
		gap: 0.375rem;
		margin: 0 0 0.3rem 0.5rem;
	}
	.gr-dmg-stepper > * {
		box-sizing: border-box;
		height: 2rem;
		line-height: 1;
		margin: 0;
	}
	.gr-dmg-btn {
		flex: 0 0 2rem;
		width: 2rem;
		border-radius: var(--gr-radius-md);
		background: var(--gr-parchment-card);
		border: 1px solid var(--gr-parchment-border-strong);
		font-size: 1.125rem;
		font-weight: 700;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
	}
	.gr-dmg-minus {
		color: var(--gr-accent);
	}
	.gr-dmg-plus {
		color: var(--gr-hp-good-2);
	}
	.gr-dmg-amount {
		flex: 0 0 auto;
		width: 2.75rem;
		border-radius: var(--gr-radius-md);
		background: var(--gr-input-bg);
		border: 1px solid var(--gr-parchment-border-strong);
		text-align: center;
		font-size: 0.875rem;
		font-weight: 700;
		color: var(--gr-ink-soft);
		font-family: var(--gr-font-body);
		padding: 0;
		-moz-appearance: textfield;
		appearance: textfield;
	}
	.gr-dmg-amount::-webkit-inner-spin-button,
	.gr-dmg-amount::-webkit-outer-spin-button {
		-webkit-appearance: none;
		margin: 0;
	}
	.gr-temp-btn {
		flex: 0 1 auto;
		min-width: 2.75rem;
		padding: 0 0.625rem;
		border-radius: var(--gr-radius-md);
		background: var(--gr-parchment-card);
		border: 1px solid var(--gr-parchment-border-strong);
		font-family: var(--gr-font-display);
		font-size: 0.625rem;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: var(--gr-player-bg);
		cursor: pointer;
	}

	.gr-empty {
		color: var(--gr-ink-muted);
		padding: 1rem;
	}

	.gr-turn-controls {
		flex: 0 0 auto;
		display: flex;
		background: var(--gr-parchment-footer);
		border-top: 1.5px solid var(--gr-parchment-border);
	}
	.gr-prev-btn {
		flex: 0 0 auto;
		font-family: var(--gr-font-body);
		font-size: 0.9375rem;
		padding: 0.7rem 0.9rem;
		background: var(--gr-parchment-footer);
		border: none;
		color: var(--gr-ink);
		cursor: pointer;
	}
	.gr-next-btn {
		flex: 1;
		font-family: var(--gr-font-display);
		font-weight: 700;
		font-size: 0.875rem;
		letter-spacing: 0.0875em;
		text-transform: uppercase;
		padding: 0.7rem;
		background: var(--gr-accent);
		box-shadow: inset 0 -3px 0 var(--gr-accent-shadow);
		border: none;
		border-radius: var(--gr-radius-lg);
		color: var(--gr-cream);
		cursor: pointer;
		margin: 0.4rem;
	}

	.gr-right {
		flex: 1;
		min-width: 0;
		background: var(--gr-parchment-panel);
		overflow-y: auto;
	}
	.gr-right-inner {
		height: 100%;
	}
	.gr-panel-msg {
		padding: 2rem;
		color: var(--gr-ink-muted);
	}
</style>
