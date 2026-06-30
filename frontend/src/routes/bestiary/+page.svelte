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

			const page = await api.get<Monster[]>(`/monsters?${params}`);
			offline = false;
			results = reset ? page : [...results, ...page];
			offset += page.length;
			hasMore = page.length === PAGE_SIZE;
			void db.monsters.bulkPut(page);
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

	function onSearchInput() {
		clearTimeout(searchTimer);
		searchTimer = setTimeout(() => runSearch(true), 300);
	}

	onMount(() => runSearch(true));
</script>

<div class="gr-bestiary">
	<h1>Справочник монстров</h1>

	<div class="filters">
		<input
			type="text"
			placeholder="Поиск по имени (ru/en)"
			bind:value={search}
			oninput={onSearchInput}
		/>
		<select bind:value={edition} onchange={() => runSearch(true)}>
			<option value="">Все редакции</option>
			<option value="2014">2014</option>
			<option value="2024">2024</option>
		</select>
	</div>

	{#if offline}
		<p class="notice">Офлайн: поиск идёт по локальному кэшу.</p>
	{/if}

	<ul class="results">
		{#each results as m (m.id)}
			<li>
				<a href={`/bestiary/${m.id}`}>
					<span class="name">{m.nameRu}</span>
					<span class="name-en">{m.nameEn}</span>
					<span class="edition-badge">{m.edition}</span>
					<span class="cr">CR {m.cr}</span>
					<span class="type">{m.type}</span>
				</a>
			</li>
		{:else}
			<li class="empty">Ничего не найдено</li>
		{/each}
	</ul>

	{#if loading}
		<p class="gr-loading-msg">Загрузка...</p>
	{:else if hasMore}
		<button class="gr-more-btn" onclick={() => runSearch(false)}>Загрузить ещё</button>
	{/if}
</div>

<style>
	.gr-bestiary {
		font-family: var(--gr-font-body);
		color: var(--gr-ink);
	}
	h1 {
		font-family: var(--gr-font-display);
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--gr-ink);
		margin: 0 0 var(--gr-space-lg);
	}
	.filters {
		display: flex;
		gap: var(--gr-space-sm);
		margin-bottom: var(--gr-space-lg);
	}
	.filters input {
		flex: 1;
		font-family: var(--gr-font-body);
		padding: 0.6rem 0.75rem;
		background: var(--gr-input-bg);
		border: 1.5px solid var(--gr-input-border);
		color: var(--gr-ink);
		border-radius: var(--gr-radius-md);
	}
	.filters select {
		font-family: var(--gr-font-body);
		padding: 0.6rem 0.75rem;
		background: var(--gr-input-bg);
		border: 1.5px solid var(--gr-input-border);
		color: var(--gr-ink);
		border-radius: var(--gr-radius-md);
	}
	.notice {
		font-size: 0.875rem;
		font-style: italic;
		color: var(--gr-ink-muted);
	}
	.results {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: var(--gr-space-xs);
	}
	.results li a {
		display: flex;
		gap: var(--gr-space-md);
		align-items: baseline;
		padding: 0.65rem 0.9rem;
		border-radius: var(--gr-radius-md);
		background: var(--gr-parchment-card);
		border: 1.5px solid var(--gr-parchment-border-strong);
		text-decoration: none;
		color: inherit;
	}
	.results li a:hover {
		background: var(--gr-parchment-highlight);
		border-color: var(--gr-accent);
	}
	.results li.empty {
		color: var(--gr-ink-muted);
		font-style: italic;
		padding: 0.5rem 0.25rem;
	}
	.name {
		font-family: var(--gr-font-display);
		font-weight: 700;
		color: var(--gr-ink);
	}
	.name-en {
		color: var(--gr-ink-faint);
		font-style: italic;
	}
	.edition-badge {
		margin-left: auto;
		font-family: var(--gr-font-display);
		font-size: 0.6875rem;
		letter-spacing: 0.04em;
		border: 1.3px solid var(--gr-parchment-border-strong);
		border-radius: 20px;
		padding: 0.1rem 0.55rem;
		color: var(--gr-ink-muted);
	}
	.cr,
	.type {
		color: var(--gr-ink-muted);
		font-size: 0.85rem;
	}
	.gr-loading-msg {
		color: var(--gr-ink-muted);
		margin-top: var(--gr-space-md);
	}
	.gr-more-btn {
		font-family: var(--gr-font-display);
		font-size: 0.75rem;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		margin-top: var(--gr-space-md);
		padding: 0.6rem 1.1rem;
		border-radius: var(--gr-radius-md);
		border: 1.3px solid var(--gr-accent);
		background: none;
		color: var(--gr-accent);
		cursor: pointer;
	}
</style>
