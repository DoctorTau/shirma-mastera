<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { db } from '$lib/db';
	import type { CreatedCreature } from '$lib/types';
	import StatBlockEditor from '$lib/components/StatBlockEditor.svelte';
	import '$lib/styles/grifel.css';

	let creature = $state<CreatedCreature | null>(null);
	let saving = $state(false);

	async function load(id: string) {
		try {
			creature = await api.get<CreatedCreature>(`/creatures/${id}`);
			void db.createdCreatures.put(creature);
		} catch {
			creature = (await db.createdCreatures.get(id)) ?? null;
		}
	}

	$effect(() => {
		if (page.params.id) load(page.params.id);
	});

	async function save() {
		if (!creature) return;
		saving = true;
		try {
			const updated = await api.put<CreatedCreature>(`/creatures/${creature.id}`, creature);
			await db.createdCreatures.put(updated);
			creature = updated;
		} finally {
			saving = false;
		}
	}

	async function duplicate() {
		if (!creature) return;
		const copy = await api.post<CreatedCreature>('/creatures', {
			nameRu: creature.nameRu + ' (копия)',
			nameEn: creature.nameEn,
			notes: creature.notes,
			statblock: creature.statblock
		});
		await db.createdCreatures.put(copy);
		await goto(`/creatures/${copy.id}`);
	}

	async function remove() {
		if (!creature || !confirm('Удалить существо?')) return;
		await api.delete(`/creatures/${creature.id}`);
		await db.createdCreatures.delete(creature.id);
		await goto('/creatures');
	}
</script>

{#if !creature}
	<p class="gr-loading-msg">Загрузка...</p>
{:else}
	<div class="gr-form-page">
		<h1>{creature.nameRu}</h1>

		<div class="form">
			<div class="row">
				<label>Имя (ru)<input bind:value={creature.nameRu} /></label>
				<label>Имя (en)<input bind:value={creature.nameEn} /></label>
			</div>
			<label>Заметки<textarea bind:value={creature.notes}></textarea></label>

			<StatBlockEditor bind:statblock={creature.statblock} />

			<div class="actions">
				<button class="gr-primary-btn" onclick={save} disabled={saving}>Сохранить</button>
				<button class="gr-outline-btn" onclick={duplicate}>Дублировать</button>
				<button class="gr-outline-btn danger" onclick={remove}>Удалить</button>
				<a href="/creatures">Назад</a>
			</div>
		</div>
	</div>
{/if}

<style>
	.gr-loading-msg {
		font-family: var(--gr-font-body);
		color: var(--gr-ink-muted);
	}
	.gr-form-page {
		font-family: var(--gr-font-body);
		color: var(--gr-ink);
	}
	h1 {
		font-family: var(--gr-font-display);
		font-size: 1.5rem;
		font-weight: 700;
		margin: 0 0 var(--gr-space-lg);
	}
	.form {
		display: flex;
		flex-direction: column;
		gap: var(--gr-space-md);
		max-width: 50rem;
	}
	.row {
		display: flex;
		gap: var(--gr-space-md);
	}
	label {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		font-size: 0.8rem;
		color: var(--gr-ink-muted);
		flex: 1;
	}
	input,
	textarea {
		font-family: var(--gr-font-body);
		padding: 0.55rem 0.7rem;
		background: var(--gr-input-bg);
		border: 1.5px solid var(--gr-input-border);
		color: var(--gr-ink);
		border-radius: var(--gr-radius-md);
	}
	.actions {
		display: flex;
		gap: var(--gr-space-md);
		align-items: center;
		flex-wrap: wrap;
	}
	.gr-primary-btn {
		font-family: var(--gr-font-display);
		font-weight: 700;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		font-size: 0.8rem;
		background: var(--gr-accent);
		box-shadow: inset 0 -2px 0 var(--gr-accent-shadow);
		color: var(--gr-cream);
		border: none;
		padding: 0.65rem 1.3rem;
		border-radius: var(--gr-radius-md);
		cursor: pointer;
	}
	.gr-primary-btn:disabled {
		opacity: 0.5;
		cursor: default;
	}
	.gr-outline-btn {
		font-family: var(--gr-font-display);
		font-size: 0.8rem;
		letter-spacing: 0.03em;
		background: none;
		border: 1.3px solid var(--gr-parchment-border-strong);
		color: var(--gr-ink-muted);
		padding: 0.65rem 1.1rem;
		border-radius: var(--gr-radius-md);
		cursor: pointer;
	}
	.gr-outline-btn.danger {
		border-color: var(--gr-accent);
		color: var(--gr-accent);
	}
	.actions a {
		color: var(--gr-ink-muted);
		font-size: 0.85rem;
		font-style: italic;
	}
</style>
