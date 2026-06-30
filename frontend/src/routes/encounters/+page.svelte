<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { db } from '$lib/db';
	import type { Encounter } from '$lib/types';
	import '$lib/styles/grifel.css';

	let encounters = $state<Encounter[]>([]);
	let loading = $state(true);
	let newName = $state('');

	async function load() {
		loading = true;
		try {
			encounters = await api.get<Encounter[]>('/encounters');
			void db.encounters.bulkPut(encounters);
		} catch {
			encounters = await db.encounters.toArray();
		} finally {
			loading = false;
		}
	}

	async function create() {
		if (!newName) return;
		const created = await api.post<Encounter>('/encounters', { name: newName });
		await db.encounters.put(created);
		newName = '';
		await goto(`/encounters/${created.id}`);
	}

	async function duplicate(id: string) {
		const copy = await api.post<Encounter>(`/encounters/${id}/duplicate`);
		await db.encounters.put(copy);
		encounters = [copy, ...encounters];
	}

	async function remove(id: string) {
		if (!confirm('Удалить энкаунтер?')) return;
		await api.delete(`/encounters/${id}`);
		await db.encounters.delete(id);
		encounters = encounters.filter((e) => e.id !== id);
	}

	onMount(load);
</script>

<div class="gr-encounters">
	<h1>Энкаунтеры</h1>

	<div class="new-row">
		<input placeholder="Название нового энкаунтера" bind:value={newName} />
		<button class="gr-create-btn" onclick={create} disabled={!newName}>Создать</button>
	</div>

	{#if loading}
		<p class="gr-loading-msg">Загрузка...</p>
	{:else}
		<ul class="list">
			{#each encounters as e (e.id)}
				<li>
					<a href={e.status === 'active' ? `/combat/${e.id}` : `/encounters/${e.id}`}>
						<span class="name">{e.name}</span>
						<span class="status status-{e.status}">{e.status}</span>
					</a>
					<div class="row-actions">
						<button onclick={() => duplicate(e.id)}>Дублировать</button>
						<button class="danger" onclick={() => remove(e.id)}>Удалить</button>
					</div>
				</li>
			{:else}
				<li class="empty">Пока нет энкаунтеров</li>
			{/each}
		</ul>
	{/if}
</div>

<style>
	.gr-encounters {
		font-family: var(--gr-font-body);
		color: var(--gr-ink);
	}
	h1 {
		font-family: var(--gr-font-display);
		font-size: 1.5rem;
		font-weight: 700;
		margin: 0 0 var(--gr-space-lg);
	}
	.new-row {
		display: flex;
		gap: var(--gr-space-sm);
		margin-bottom: var(--gr-space-lg);
	}
	.new-row input {
		flex: 1;
		font-family: var(--gr-font-body);
		padding: 0.6rem 0.75rem;
		background: var(--gr-input-bg);
		border: 1.5px solid var(--gr-input-border);
		color: var(--gr-ink);
		border-radius: var(--gr-radius-md);
	}
	.gr-create-btn {
		font-family: var(--gr-font-display);
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		background: var(--gr-accent);
		box-shadow: inset 0 -2px 0 var(--gr-accent-shadow);
		color: var(--gr-cream);
		border: none;
		padding: 0 var(--gr-space-lg);
		border-radius: var(--gr-radius-md);
		cursor: pointer;
	}
	.gr-create-btn:disabled {
		opacity: 0.5;
		cursor: default;
	}
	.list {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: var(--gr-space-xs);
	}
	.list li {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: var(--gr-space-md);
		padding: 0.7rem 1rem;
		border-radius: var(--gr-radius-md);
		background: var(--gr-parchment-card);
		border: 1.5px solid var(--gr-parchment-border-strong);
	}
	.list li.empty {
		color: var(--gr-ink-muted);
		font-style: italic;
		border: none;
		background: none;
		padding: 0.5rem 0.25rem;
	}
	.list a {
		display: flex;
		gap: var(--gr-space-md);
		align-items: center;
		color: inherit;
		text-decoration: none;
		min-width: 0;
	}
	.name {
		font-family: var(--gr-font-display);
		font-weight: 700;
	}
	.status {
		font-family: var(--gr-font-display);
		font-size: 0.6875rem;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		border: 1.3px solid var(--gr-parchment-border-strong);
		border-radius: 20px;
		padding: 0.1rem 0.6rem;
		color: var(--gr-ink-muted);
	}
	.status-active {
		border-color: var(--gr-accent);
		color: var(--gr-accent);
		background: var(--gr-tag-warn-bg);
	}
	.row-actions {
		display: flex;
		gap: var(--gr-space-sm);
		flex: none;
	}
	.row-actions button {
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
	.row-actions button.danger {
		border-color: var(--gr-accent);
		color: var(--gr-accent);
	}
</style>
