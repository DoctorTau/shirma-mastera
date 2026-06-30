<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { db } from '$lib/db';
	import type { CreatedCreature, PlayerCharacter } from '$lib/types';
	import '$lib/styles/grifel.css';

	let creatures = $state<CreatedCreature[]>([]);
	let players = $state<PlayerCharacter[]>([]);
	let loading = $state(true);

	async function load() {
		loading = true;
		try {
			creatures = await api.get<CreatedCreature[]>('/creatures');
			players = await api.get<PlayerCharacter[]>('/players');
			void db.createdCreatures.bulkPut(creatures);
			void db.playerCharacters.bulkPut(players);
		} catch {
			creatures = await db.createdCreatures.toArray();
			players = await db.playerCharacters.toArray();
		} finally {
			loading = false;
		}
	}

	async function deleteCreature(id: string) {
		if (!confirm('Удалить существо?')) return;
		await api.delete(`/creatures/${id}`);
		await db.createdCreatures.delete(id);
		creatures = creatures.filter((c) => c.id !== id);
	}

	async function deletePlayer(id: string) {
		if (!confirm('Удалить игрока?')) return;
		await api.delete(`/players/${id}`);
		await db.playerCharacters.delete(id);
		players = players.filter((p) => p.id !== id);
	}

	onMount(load);
</script>

<div class="gr-creatures">
	<h1>Мои существа и игроки</h1>

	{#if loading}
		<p class="gr-loading-msg">Загрузка...</p>
	{:else}
		<section>
			<div class="section-header">
				<h2>Существа и NPC</h2>
				<a class="btn" href="/creatures/new">+ Новое существо</a>
			</div>
			<ul class="list">
				{#each creatures as c (c.id)}
					<li>
						<a href={`/creatures/${c.id}`}>{c.nameRu} <span class="name-en">{c.nameEn}</span></a>
						<button onclick={() => deleteCreature(c.id)}>Удалить</button>
					</li>
				{:else}
					<li class="empty">Пока нет своих существ</li>
				{/each}
			</ul>
		</section>

		<section>
			<div class="section-header">
				<h2>Игроки</h2>
				<a class="btn" href="/players/new">+ Новый игрок</a>
			</div>
			<ul class="list">
				{#each players as p (p.id)}
					<li>
						<a href={`/players/${p.id}`}>{p.name}</a>
						<button onclick={() => deletePlayer(p.id)}>Удалить</button>
					</li>
				{:else}
					<li class="empty">Пока нет игроков</li>
				{/each}
			</ul>
		</section>
	{/if}
</div>

<style>
	.gr-creatures {
		font-family: var(--gr-font-body);
		color: var(--gr-ink);
	}
	h1 {
		font-family: var(--gr-font-display);
		font-size: 1.5rem;
		font-weight: 700;
		margin: 0 0 var(--gr-space-lg);
	}
	.gr-loading-msg {
		color: var(--gr-ink-muted);
	}
	section {
		margin-bottom: var(--gr-space-xl);
	}
	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--gr-space-sm);
	}
	h2 {
		font-family: var(--gr-font-display);
		font-size: 0.8rem;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--gr-accent);
		border-bottom: 1.5px solid var(--gr-accent);
		padding-bottom: 0.3rem;
		margin: 0;
		flex: 1;
	}
	.btn {
		font-family: var(--gr-font-display);
		font-size: 0.6875rem;
		font-weight: 600;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		background: var(--gr-accent);
		color: var(--gr-cream);
		padding: 0.5rem 0.9rem;
		border-radius: var(--gr-radius-md);
		text-decoration: none;
		white-space: nowrap;
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
		padding: 0.6rem 0.9rem;
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
		font-family: var(--gr-font-display);
		font-weight: 700;
		color: var(--gr-ink);
		text-decoration: none;
	}
	.name-en {
		font-family: var(--gr-font-body);
		font-weight: 400;
		color: var(--gr-ink-faint);
		font-style: italic;
	}
	.list button {
		font-family: var(--gr-font-display);
		font-size: 0.6875rem;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		background: none;
		border: 1.3px solid var(--gr-accent);
		color: var(--gr-accent);
		border-radius: var(--gr-radius-sm);
		padding: 0.35rem 0.65rem;
		cursor: pointer;
	}
</style>
