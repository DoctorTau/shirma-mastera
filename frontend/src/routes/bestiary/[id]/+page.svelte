<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { db } from '$lib/db';
	import type { Monster, CreatedCreature } from '$lib/types';
	import GrifelStatBlock from '$lib/components/GrifelStatBlock.svelte';
	import '$lib/styles/grifel.css';

	let monster = $state<Monster | null>(null);
	let linked = $state<Monster | null>(null);
	let loading = $state(true);
	let notFound = $state(false);

	async function load(id: string) {
		loading = true;
		notFound = false;
		linked = null;
		try {
			monster = await api.get<Monster>(`/monsters/${id}`);
			void db.monsters.put(monster);
		} catch {
			monster = (await db.monsters.get(id)) ?? null;
			if (!monster) notFound = true;
		}
		if (monster?.linkedMonsterId) {
			try {
				linked = await api.get<Monster>(`/monsters/${monster.linkedMonsterId}`);
			} catch {
				linked = (await db.monsters.get(monster.linkedMonsterId)) ?? null;
			}
		}
		loading = false;
	}

	$effect(() => {
		if (page.params.id) load(page.params.id);
	});

	async function createFromMonster() {
		if (!monster) return;
		const created = await api.post<CreatedCreature>('/creatures', {
			nameRu: monster.nameRu,
			nameEn: monster.nameEn,
			notes: '',
			statblock: monster.statblock
		});
		await db.createdCreatures.put(created);
		await goto(`/creatures/${created.id}`);
	}
</script>

{#if loading}
	<p class="gr-loading-msg">Загрузка...</p>
{:else if notFound || !monster}
	<p class="gr-loading-msg">Монстр не найден (нет сети и нет в офлайн-кэше).</p>
{:else}
	<div class="gr-detail-bar">
		<div class="edition-toggle">
			<span class="label">Редакция:</span>
			{#each ['2014', '2024'] as ed (ed)}
				{#if ed === monster.edition}
					<button class="active" disabled>{ed}</button>
				{:else}
					<button disabled={!linked} onclick={() => linked && goto(`/bestiary/${linked.id}`)}>
						{linked ? ed : `${ed} — нет`}
					</button>
				{/if}
			{/each}
		</div>

		<button class="create-from" onclick={createFromMonster}>Создать своего на основе</button>
	</div>

	<div class="gr-statblock-wrap">
		<GrifelStatBlock
			statblock={monster.statblock}
			nameRu={monster.nameRu}
			nameEn={monster.nameEn}
			sourceUrl={monster.sourceUrl}
			imageUrl={monster.imageUrl}
		/>
	</div>
{/if}

<style>
	.gr-loading-msg {
		font-family: var(--gr-font-body);
		color: var(--gr-ink-muted);
	}
	.gr-detail-bar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-wrap: wrap;
		gap: var(--gr-space-md);
		margin-bottom: var(--gr-space-lg);
	}
	.edition-toggle {
		display: flex;
		align-items: center;
		gap: var(--gr-space-sm);
		font-family: var(--gr-font-body);
	}
	.label {
		color: var(--gr-ink-muted);
		font-size: 0.875rem;
	}
	.edition-toggle button {
		font-family: var(--gr-font-display);
		font-size: 0.75rem;
		letter-spacing: 0.04em;
		padding: 0.4rem 0.9rem;
		border-radius: var(--gr-radius-md);
		border: 1.3px solid var(--gr-parchment-border-strong);
		background: var(--gr-parchment-card);
		color: var(--gr-ink-muted);
		cursor: pointer;
	}
	.edition-toggle button.active {
		background: var(--gr-accent);
		border-color: var(--gr-accent);
		color: var(--gr-cream);
	}
	.edition-toggle button:disabled {
		cursor: default;
	}
	.create-from {
		font-family: var(--gr-font-display);
		font-size: 0.6875rem;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		font-weight: 600;
		padding: 0.5rem 0.9rem;
		border-radius: var(--gr-radius-md);
		border: 1.3px solid var(--gr-accent);
		background: none;
		color: var(--gr-accent);
		cursor: pointer;
	}
	.gr-statblock-wrap {
		max-width: 42rem;
		background: var(--gr-parchment-card);
		border: 1.5px solid var(--gr-parchment-border-strong);
		border-radius: var(--gr-radius-xl);
		overflow: hidden;
	}
</style>
