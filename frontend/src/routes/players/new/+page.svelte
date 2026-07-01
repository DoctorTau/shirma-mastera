<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import { db } from '$lib/db';
	import type { PlayerCharacter } from '$lib/types';
	import '$lib/styles/grifel.css';

	let name = $state('');
	let ac = $state<number | undefined>(undefined);
	let passivePerception = $state<number | undefined>(undefined);
	let maxHp = $state<number | undefined>(undefined);
	let notes = $state('');
	let saving = $state(false);

	async function save() {
		saving = true;
		try {
			const created = await api.post<PlayerCharacter>('/players', {
				name,
				ac,
				passivePerception,
				maxHp,
				notes
			});
			await db.playerCharacters.put(created);
			await goto('/creatures');
		} finally {
			saving = false;
		}
	}
</script>

<div class="gr-form-page">
	<h1>Новый игрок</h1>

	<div class="form">
		<label>Имя<input bind:value={name} /></label>
		<div class="row">
			<label>КД<input type="number" bind:value={ac} /></label>
			<label>Пасс. внимательность<input type="number" bind:value={passivePerception} /></label>
			<label>Макс. хиты<input type="number" bind:value={maxHp} /></label>
		</div>
		<label>Заметки<textarea bind:value={notes}></textarea></label>

		<div class="actions">
			<button class="gr-primary-btn" onclick={save} disabled={saving || !name}>Сохранить</button>
			<a href="/creatures">Отмена</a>
		</div>
	</div>
</div>

<style>
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
		max-width: 30rem;
	}
	.row {
		display: flex;
		gap: var(--gr-space-md);
		flex-wrap: wrap;
	}
	label {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		font-size: 0.8rem;
		color: var(--gr-ink-muted);
		flex: 1;
	}
	.row label {
		flex: 1 1 8rem;
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
	.actions a {
		color: var(--gr-ink-muted);
		font-size: 0.85rem;
		font-style: italic;
	}
</style>
