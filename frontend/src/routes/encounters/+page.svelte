<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { db } from '$lib/db';
	import type { Encounter } from '$lib/types';
	import '$lib/styles/grifel.css';

	let loading = $state(true);
	let hasEncounters = $state(false);

	async function createAndEnter() {
		const created = await api.post<Encounter>('/encounters', { name: 'Новый энкаунтер' });
		await db.encounters.put(created);
		await goto(`/encounters/${created.id}`);
	}

	onMount(async () => {
		let list: Encounter[] = [];
		try {
			list = (await api.get<Encounter[]>('/encounters')) ?? [];
		} catch {
			list = await db.encounters.toArray();
		}
		if (list.length > 0) {
			const active = list.find((e) => e.status === 'active');
			await goto(`/encounters/${(active ?? list[0]).id}`, { replaceState: true });
			return;
		}
		hasEncounters = list.length > 0;
		loading = false;
	});
</script>

{#if loading}
	<p class="gr-loading-msg">Загрузка...</p>
{:else if !hasEncounters}
	<div class="gr-empty-state">
		<div class="cz">Пока нет энкаунтеров</div>
		<p>Соберите первую стычку: добавьте монстров и игроков, затем запустите бой.</p>
		<button class="gr-start-btn" onclick={createAndEnter}>+ Новый энкаунтер</button>
	</div>
{/if}

<style>
	.gr-loading-msg {
		font-family: var(--gr-font-body);
		color: var(--gr-ink-muted);
	}
	.gr-empty-state {
		max-width: 28rem;
		margin: var(--gr-space-3xl) auto 0;
		text-align: center;
		font-family: var(--gr-font-body);
		color: var(--gr-ink);
	}
	.gr-empty-state .cz {
		font-family: var(--gr-font-display);
		font-size: 1.25rem;
		font-weight: 700;
		color: var(--gr-ink);
		margin-bottom: var(--gr-space-sm);
	}
	.gr-empty-state p {
		color: var(--gr-ink-muted);
		margin: 0 0 var(--gr-space-lg);
	}
	.gr-start-btn {
		font-family: var(--gr-font-display);
		font-size: 0.8125rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		background: var(--gr-accent);
		box-shadow: inset 0 -3px 0 var(--gr-accent-shadow);
		color: var(--gr-cream);
		border: none;
		border-radius: var(--gr-radius-lg);
		padding: 0.75rem 1.5rem;
		cursor: pointer;
	}
</style>
