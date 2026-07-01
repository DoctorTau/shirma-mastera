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
		{ href: '/encounters', label: 'Бой', icon: '⚔' },
		{ href: '/creatures', label: 'Мои существа', icon: '✎' },
		{ href: '/bestiary', label: 'Справочник', icon: '▤' }
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

{#if $authToken}
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
						class:gr-tab-active={page.url.pathname.startsWith(item.href) ||
							(item.href === '/encounters' && page.url.pathname.startsWith('/combat/'))}
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
		<nav class="gr-bottom-tabs">
			{#each navItems as item (item.href)}
				<a
					href={item.href}
					class="gr-bottom-tab"
					class:gr-bottom-tab-active={page.url.pathname.startsWith(item.href) ||
						(item.href === '/encounters' && page.url.pathname.startsWith('/combat/'))}
				>
					<span class="gr-bottom-tab-icon">{item.icon}</span>
					<span class="gr-bottom-tab-label">{item.label}</span>
				</a>
			{/each}
		</nav>
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

	.gr-bottom-tabs {
		display: none;
	}

	/* Phone-width chrome: top tab links collapse into a fixed bottom tab
	   bar (PWA convention — thumb-reachable, matches the iPhone reference). */
	@media (max-width: 680px) {
		.gr-tabs {
			display: none;
		}
		.gr-header-spacer {
			flex: 1;
		}
		.app-main {
			padding: var(--gr-space-lg);
			padding-bottom: calc(var(--gr-space-lg) + 62px + env(safe-area-inset-bottom, 0px));
		}
		.gr-bottom-tabs {
			position: fixed;
			left: 0;
			right: 0;
			bottom: 0;
			z-index: 20;
			display: flex;
			gap: 4px;
			padding: 8px 10px calc(8px + env(safe-area-inset-bottom, 10px));
			background: #1a1310;
			background-image: linear-gradient(180deg, #241a14, #160f0b);
			border-top: 1px solid var(--gr-ink-soft);
		}
		.gr-bottom-tab {
			flex: 1;
			display: flex;
			flex-direction: column;
			align-items: center;
			gap: 3px;
			padding: 5px 0;
			color: var(--gr-cream-dim);
			opacity: 0.68;
			text-decoration: none;
			transition:
				color var(--gr-duration-base) var(--gr-ease),
				opacity var(--gr-duration-base) var(--gr-ease);
		}
		.gr-bottom-tab-icon {
			font-size: 1.1875rem;
			line-height: 1;
		}
		.gr-bottom-tab-label {
			font-family: var(--gr-font-display);
			font-size: 0.59375rem;
			letter-spacing: 0.06em;
			text-transform: uppercase;
		}
		.gr-bottom-tab-active {
			color: var(--gr-cream);
			opacity: 1;
		}
		.gr-bottom-tab-active .gr-bottom-tab-label {
			font-weight: 700;
		}
	}
</style>
