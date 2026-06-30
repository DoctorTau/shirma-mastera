<script lang="ts">
	import { goto } from '$app/navigation';
	import { login, ApiError } from '$lib/api';
	import '$lib/styles/grifel.css';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	async function submit(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;
		try {
			await login(email, password);
			await goto('/encounters');
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Не удалось войти';
		} finally {
			loading = false;
		}
	}
</script>

<div class="gr-page login-screen">
	<form onsubmit={submit}>
		<div class="gr-brand">
			<span class="gr-brand-mark">✦</span>
			<h1>Грифель</h1>
			<p class="gr-tagline">DM-компаньон</p>
		</div>

		<!-- svelte-ignore a11y_autofocus -->
		<input type="email" bind:value={email} placeholder="Email" autofocus />
		<input type="password" bind:value={password} placeholder="Пароль" />
		{#if error}
			<p class="error">{error}</p>
		{/if}
		<button type="submit" disabled={loading || !email || !password}>Войти</button>
		<a href="/register">Создать аккаунт</a>
	</form>
</div>

<style>
	.login-screen {
		display: flex;
		align-items: center;
		justify-content: center;
	}
	form {
		display: flex;
		flex-direction: column;
		gap: var(--gr-space-md);
		width: 22rem;
		max-width: calc(100vw - 2.5rem);
		background: var(--gr-parchment-panel);
		border: 1px solid var(--gr-parchment-border);
		border-radius: var(--gr-radius-xl);
		padding: var(--gr-space-2xl) var(--gr-space-xl);
		box-shadow: 0 30px 70px -25px rgba(0, 0, 0, 0.7);
	}
	.gr-brand {
		text-align: center;
		margin-bottom: var(--gr-space-sm);
	}
	.gr-brand-mark {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 2.5rem;
		height: 2.5rem;
		border: 1.5px solid var(--gr-accent);
		border-radius: 50%;
		color: var(--gr-accent);
		font-size: 1.1rem;
		margin-bottom: var(--gr-space-sm);
	}
	h1 {
		font-family: var(--gr-font-display);
		font-size: 1.6rem;
		font-weight: 700;
		letter-spacing: 0.04em;
		color: var(--gr-ink);
		margin: 0;
	}
	.gr-tagline {
		margin: 0.2rem 0 0;
		font-size: 0.8rem;
		font-style: italic;
		color: var(--gr-ink-muted);
	}
	input {
		font-family: var(--gr-font-body);
		padding: 0.7rem 0.85rem;
		border-radius: var(--gr-radius-md);
		border: 1.5px solid var(--gr-input-border);
		background: var(--gr-input-bg);
		color: var(--gr-ink);
		font-size: 1rem;
	}
	input::placeholder {
		color: var(--gr-ink-faint);
	}
	button {
		font-family: var(--gr-font-display);
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		padding: 0.8rem;
		border-radius: var(--gr-radius-lg);
		border: none;
		background: var(--gr-accent);
		box-shadow: inset 0 -3px 0 var(--gr-accent-shadow);
		color: var(--gr-cream);
		font-size: 0.875rem;
		cursor: pointer;
		margin-top: var(--gr-space-xs);
	}
	button:disabled {
		opacity: 0.5;
		cursor: default;
	}
	.error {
		color: var(--gr-accent);
		margin: 0;
		font-size: 0.85rem;
	}
	a {
		text-align: center;
		color: var(--gr-ink-muted);
		font-size: 0.85rem;
		font-style: italic;
	}
</style>
