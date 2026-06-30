<script lang="ts">
	import favicon from '$lib/assets/favicon.svg';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { authToken } from '$lib/stores/auth';
	import { requestPersistentStorage } from '$lib/db';
	import { startSyncListeners } from '$lib/sync';
	import '$lib/styles/grifel.css';

	let { children } = $props();

	const navItems = [
		{ href: '/encounters', label: 'Бой' },
		{ href: '/creatures', label: 'Мои существа' },
		{ href: '/bestiary', label: 'Справочник' }
	];

	onMount(() => {
		void requestPersistentStorage();
		startSyncListeners();
	});

	$effect(() => {
		const isAuthPage = page.url.pathname === '/login' || page.url.pathname === '/register';
		if (!$authToken && !isAuthPage) {
			goto('/login');
		}
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{#if $authToken && page.url.pathname.startsWith('/combat/')}
	{@render children()}
{:else if $authToken}
	<div class="app-shell">
		<header class="gr-header">
			<div class="gr-logo">
				<span class="gr-logo-mark">✦</span>
				Грифель
			</div>
			<nav class="gr-tabs">
				{#each navItems as item (item.href)}
					<a
						href={item.href}
						class="gr-tab"
						class:gr-tab-active={page.url.pathname.startsWith(item.href)}
					>
						{item.label}
					</a>
				{/each}
			</nav>
			<div class="gr-header-spacer"></div>
			<button class="gr-logout" onclick={() => authToken.clear()}>Выйти</button>
		</header>
		<main class="app-main">
			{@render children()}
		</main>
	</div>
{:else}
	{@render children()}
{/if}

<style>
	.app-shell {
		display: flex;
		flex-direction: column;
		height: 100vh;
		font-family: var(--gr-font-body);
		color: var(--gr-ink);
	}
	.gr-header {
		flex: 0 0 auto;
		height: 54px;
		display: flex;
		align-items: center;
		gap: var(--gr-space-lg);
		padding: 0 var(--gr-space-lg);
		background: linear-gradient(var(--gr-maroon), var(--gr-maroon-dark));
		border-bottom: 2px solid var(--gr-maroon-deep);
		color: var(--gr-cream);
	}
	.gr-logo {
		font-family: var(--gr-font-display);
		font-weight: 700;
		font-size: 0.9375rem;
		letter-spacing: 0.15em;
		color: var(--gr-cream);
		white-space: nowrap;
		display: flex;
		align-items: center;
		gap: var(--gr-space-sm);
	}
	.gr-logo-mark {
		width: 22px;
		height: 22px;
		border: 1.5px solid var(--gr-cream-dim);
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.6875rem;
	}
	.gr-tabs {
		display: flex;
		gap: 0.25rem;
	}
	.gr-tab {
		font-family: var(--gr-font-display);
		font-size: 0.75rem;
		font-weight: 400;
		letter-spacing: 0.075em;
		text-transform: uppercase;
		color: var(--gr-cream-soft);
		opacity: 0.72;
		padding: 0.4375rem 0.875rem;
		border-radius: var(--gr-radius-md);
		text-decoration: none;
	}
	.gr-tab-active {
		background: var(--gr-parchment-panel);
		color: var(--gr-accent);
		font-weight: 700;
		opacity: 1;
		box-shadow: 0 1px 0 rgba(0, 0, 0, 0.2);
	}
	.gr-header-spacer {
		flex: 1;
	}
	.gr-logout {
		font-family: var(--gr-font-display);
		font-size: 0.6875rem;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--gr-cream-soft);
		background: none;
		border: 1px solid rgba(216, 184, 136, 0.4);
		border-radius: var(--gr-radius-md);
		padding: 0.4375rem 0.75rem;
		cursor: pointer;
	}
	.gr-logout:hover {
		border-color: var(--gr-cream-dim);
		color: var(--gr-cream);
	}
	.app-main {
		flex: 1;
		overflow: auto;
		background: var(--gr-parchment-panel);
		background-image: radial-gradient(circle at 20% 0%, rgba(255, 255, 255, 0.5), transparent 60%);
		padding: var(--gr-space-xl);
	}
</style>
