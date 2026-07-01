<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { db } from '$lib/db';
	import type { Monster, Edition } from '$lib/types';
	import '$lib/styles/grifel.css';

	let search = $state('');
	let edition = $state<'' | Edition>('');
	let results = $state<Monster[]>([]);
	let loading = $state(false);
	let offline = $state(false);
	let offset = $state(0);
	let hasMore = $state(true);
	let cachedCount = $state(0);
	const PAGE_SIZE = 100;

	let searchTimer: ReturnType<typeof setTimeout>;

	async function runSearch(reset: boolean) {
		if (reset) {
			offset = 0;
			results = [];
			hasMore = true;
		}
		loading = true;
		try {
			const params = new URLSearchParams();
			if (search) params.set('search', search);
			if (edition) params.set('edition', edition);
			params.set('limit', String(PAGE_SIZE));
			params.set('offset', String(offset));

			const page = (await api.get<Monster[]>(`/monsters?${params}`)) ?? [];
			offline = false;
			results = reset ? page : [...results, ...page];
			offset += page.length;
			hasMore = page.length === PAGE_SIZE;
			void db.monsters.bulkPut(page).then(refreshCacheCount);
		} catch {
			offline = true;
			results = await searchOffline();
			hasMore = false;
		} finally {
			loading = false;
		}
	}

	async function searchOffline(): Promise<Monster[]> {
		const all = await db.monsters.toArray();
		return all.filter((m) => {
			if (edition && m.edition !== edition) return false;
			if (!search) return true;
			const q = search.toLowerCase();
			return m.nameRu.toLowerCase().includes(q) || m.nameEn.toLowerCase().includes(q);
		});
	}

	async function refreshCacheCount() {
		cachedCount = await db.monsters.count();
	}

	function onSearchInput() {
		clearTimeout(searchTimer);
		searchTimer = setTimeout(() => runSearch(true), 300);
	}

	function setEdition(next: '' | Edition) {
		edition = edition === next ? '' : next;
		runSearch(true);
	}

	onMount(() => {
		runSearch(true);
		void refreshCacheCount();
	});
</script>

<div class="gr-bestiary">
	<aside class="rail">
		<div class="rail-label">Разделы dnd.su</div>
		<div class="rail-item rail-item-active">
			Бестиарий
			{#if results.length}<span class="rail-count">{results.length}{hasMore ? '+' : ''}</span>{/if}
		</div>
		<div class="rail-item rail-item-soon">Заклинания<span class="rail-soon">скоро</span></div>
		<div class="rail-item rail-item-soon">Магические предметы<span class="rail-soon">скоро</span></div>
		<div class="rail-item rail-item-soon">Классы<span class="rail-soon">скоро</span></div>
		<div class="rail-item rail-item-soon">Расы и происхождения<span class="rail-soon">скоро</span></div>
		<div class="rail-item rail-item-soon">Предыстории<span class="rail-soon">скоро</span></div>
		<div class="rail-item rail-item-soon">Черты<span class="rail-soon">скоро</span></div>

		<div class="cache-card">
			<div class="cache-label">Офлайн-кэш</div>
			{#if cachedCount > 0}
				В кэше {cachedCount} {cachedCount === 1 ? 'существо' : 'существ'} · доступно без сети
			{:else}
				Кэш пуст — откройте бестиарий в сети хотя бы раз
			{/if}
		</div>
	</aside>

	<div class="main">
		<div class="toolbar">
			<input
				class="search-input"
				type="text"
				placeholder="Поиск по бестиарию (ru/en)…"
				bind:value={search}
				oninput={onSearchInput}
			/>
			<div class="pill-group">
				<button class="pill" class:pill-on={edition === '2024'} onclick={() => setEdition('2024')}
					>Редакция 2024</button
				>
				<button class="pill" class:pill-on={edition === '2014'} onclick={() => setEdition('2014')}
					>Редакция 2014</button
				>
			</div>
			<span class="count">{results.length} существ{hasMore ? '+' : ''}</span>
		</div>

		{#if offline}
			<p class="notice">Офлайн: поиск идёт по локальному кэшу.</p>
		{/if}

		<div class="grid">
			{#each results as m (m.id)}
				<a class="tile" href={`/bestiary/${m.id}`}>
					<div class="tile-top">
						<div class="tile-name">{m.nameRu}</div>
						<div class="tile-sub">{m.nameEn} · {m.type}</div>
					</div>
					<div class="tile-bottom">
						<span class="tile-cr">CR {m.cr}</span>
						<span class="tile-size">{m.size}</span>
					</div>
				</a>
			{:else}
				{#if !loading}
					<p class="empty">Ничего не найдено</p>
				{/if}
			{/each}
		</div>

		{#if loading}
			<p class="gr-loading-msg">Загрузка...</p>
		{:else if hasMore}
			<button class="gr-more-btn" onclick={() => runSearch(false)}>Загрузить ещё</button>
		{/if}
	</div>
</div>

<style>
	.gr-bestiary {
		display: flex;
		gap: 0;
		margin: calc(var(--gr-space-xl) * -1);
		height: calc(100vh - 54px);
		font-family: var(--gr-font-body);
		color: var(--gr-ink);
	}

	.rail {
		width: 210px;
		flex: none;
		overflow-y: auto;
		background: var(--gr-parchment-footer);
		border-right: 1.5px solid var(--gr-parchment-border);
		padding: var(--gr-space-lg) var(--gr-space-md);
		display: flex;
		flex-direction: column;
		gap: 5px;
	}
	.rail-label {
		font-family: var(--gr-font-display);
		font-size: 0.625rem;
		letter-spacing: 0.16em;
		text-transform: uppercase;
		color: var(--gr-ink-faint);
		padding: 4px 8px 7px;
	}
	.rail-item {
		font-size: 0.8125rem;
		padding: 0.5625rem 0.75rem;
		border-radius: var(--gr-radius-lg);
		color: var(--gr-ink-soft);
		display: flex;
		align-items: center;
		justify-content: space-between;
	}
	.rail-item-active {
		font-family: var(--gr-font-display);
		font-weight: 600;
		background: var(--gr-accent);
		color: var(--gr-cream);
	}
	.rail-count {
		opacity: 0.75;
	}
	.rail-item-soon {
		color: var(--gr-ink-faint);
		cursor: default;
	}
	.rail-soon {
		font-family: var(--gr-font-display);
		font-size: 0.5625rem;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		opacity: 0.6;
	}
	.cache-card {
		margin-top: auto;
		background: var(--gr-parchment-panel);
		border: 1.3px solid var(--gr-parchment-border);
		border-radius: var(--gr-radius-lg);
		padding: 0.625rem;
		font-size: 0.6875rem;
		color: var(--gr-ink-muted);
		line-height: 1.5;
	}
	.cache-label {
		font-family: var(--gr-font-display);
		font-size: 0.5625rem;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: var(--gr-accent);
		margin-bottom: 3px;
	}

	.main {
		flex: 1;
		min-width: 0;
		min-height: 0;
		display: flex;
		flex-direction: column;
	}

	.toolbar {
		flex: none;
		display: flex;
		align-items: center;
		gap: var(--gr-space-sm);
		padding: 0.6875rem var(--gr-space-lg);
		border-bottom: 1.5px solid var(--gr-parchment-border);
		background: var(--gr-parchment-list);
		flex-wrap: wrap;
	}
	.search-input {
		font-family: var(--gr-font-body);
		font-style: italic;
		padding: 0.4375rem 0.8rem;
		background: var(--gr-input-bg);
		border: 1.5px solid var(--gr-input-border);
		color: var(--gr-ink);
		border-radius: var(--gr-radius-lg);
		min-width: 16rem;
		flex: 1;
		max-width: 22rem;
	}
	.pill-group {
		display: flex;
		gap: var(--gr-space-xs);
	}
	.pill {
		font-family: var(--gr-font-display);
		font-size: 0.6875rem;
		letter-spacing: 0.04em;
		padding: 0.3125rem 0.6875rem;
		border-radius: 20px;
		border: 1.3px solid var(--gr-parchment-border-strong);
		background: none;
		color: var(--gr-ink-soft);
		cursor: pointer;
	}
	.pill-on {
		background: var(--gr-accent);
		border-color: var(--gr-accent);
		color: var(--gr-cream);
	}
	.count {
		margin-left: auto;
		font-size: 0.8125rem;
		color: var(--gr-ink-muted);
		white-space: nowrap;
	}

	.notice {
		margin: var(--gr-space-sm) var(--gr-space-lg) 0;
		font-size: 0.875rem;
		font-style: italic;
		color: var(--gr-ink-muted);
	}

	.grid {
		flex: 1;
		overflow: auto;
		padding: var(--gr-space-md) var(--gr-space-lg);
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
		gap: var(--gr-space-sm);
		align-content: start;
	}
	.tile {
		display: flex;
		flex-direction: column;
		justify-content: space-between;
		gap: var(--gr-space-sm);
		min-height: 88px;
		padding: 0.6875rem 0.8125rem;
		border-radius: var(--gr-radius-lg);
		background: var(--gr-parchment-card);
		border: 1.5px solid var(--gr-parchment-border-strong);
		text-decoration: none;
		color: inherit;
	}
	@media (hover: hover) and (pointer: fine) {
		.tile:hover {
			background: var(--gr-parchment-highlight);
			border-color: var(--gr-accent);
		}
	}
	.tile-name {
		font-family: var(--gr-font-display);
		font-weight: 700;
		font-size: 0.9375rem;
		color: var(--gr-ink);
	}
	.tile-sub {
		font-size: 0.6875rem;
		font-style: italic;
		color: var(--gr-ink-faint);
		margin-top: 0.15rem;
	}
	.tile-bottom {
		display: flex;
		justify-content: space-between;
		font-size: 0.6875rem;
		color: var(--gr-accent);
	}
	.tile-size {
		color: var(--gr-ink-muted);
	}
	.empty {
		grid-column: 1 / -1;
		color: var(--gr-ink-muted);
		font-style: italic;
		padding: var(--gr-space-md) 0;
	}

	.gr-loading-msg {
		flex: none;
		color: var(--gr-ink-muted);
		padding: 0 var(--gr-space-lg) var(--gr-space-md);
	}
	.gr-more-btn {
		flex: none;
		align-self: center;
		margin: 0 0 var(--gr-space-lg);
		font-family: var(--gr-font-display);
		font-size: 0.75rem;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		padding: 0.6rem 1.1rem;
		border-radius: var(--gr-radius-md);
		border: 1.3px solid var(--gr-accent);
		background: none;
		color: var(--gr-accent);
		cursor: pointer;
	}

	/* Phone width: the section rail has no room as a sidebar, so it becomes
	   a horizontal-scroll chip strip above the results, and the tile grid
	   drops to a single column. */
	@media (max-width: 680px) {
		.gr-bestiary {
			flex-direction: column;
			margin: calc(var(--gr-space-lg) * -1);
		}
		.rail {
			width: 100%;
			flex-direction: row;
			align-items: center;
			overflow-x: auto;
			overflow-y: visible;
			border-right: none;
			border-bottom: 1.5px solid var(--gr-parchment-border);
			padding: var(--gr-space-sm) var(--gr-space-md);
		}
		.rail-label {
			display: none;
		}
		.rail-item {
			flex: none;
			white-space: nowrap;
			border-radius: 20px;
			padding: 0.4375rem 0.8125rem;
			font-size: 0.75rem;
		}
		.cache-card {
			display: none;
		}
		.grid {
			grid-template-columns: 1fr;
			padding-bottom: calc(var(--gr-space-lg) + 62px + env(safe-area-inset-bottom, 0px));
		}
	}
</style>
