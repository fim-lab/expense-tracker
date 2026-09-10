<script lang="ts">
	import { page } from '$app/state';
	import { browser } from '$app/environment';
	import SunIcon from '$lib/components/icons/SunIcon.svelte';
	import MoonIcon from '$lib/components/icons/MoonIcon.svelte';

	async function handleLogout() {
		const res = await fetch('/auth/logout', { method: 'POST' });

		if (res.ok && browser) {
			window.location.href = '/login';
		}
	}

	let isDarkMode = $state(false);

	if (browser) {
		const storedTheme = document.documentElement.getAttribute('data-theme');
		isDarkMode =
			storedTheme === 'dark' ||
			(!storedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches);
	}

	function toggleTheme() {
		isDarkMode = !isDarkMode;
		const theme = isDarkMode ? 'dark' : 'light';
		document.documentElement.setAttribute('data-theme', theme);
		localStorage.setItem('theme', theme);
	}

	let { children } = $props();
</script>

{#if page.url.pathname !== '/login'}
	<nav class="container">
		<ul><li><strong>FinanceTracker</strong></li></ul>
		<ul>
			<li><a href="/">Dashboard</a></li>
			<li><a href="/transactions/add">Add TX</a></li>
			<li><a href="/wallets/transfer">Transfer</a></li>
			<li><a href="/depots">Depots</a></li>
			<li><a href="/settings">Settings</a></li>
			<li>
				<button
					class="theme-toggle"
					onclick={toggleTheme}
					aria-label={isDarkMode ? 'Switch to light mode' : 'Switch to dark mode'}
					title={isDarkMode ? 'Switch to light mode' : 'Switch to dark mode'}
				>
					{#if isDarkMode}
						<SunIcon />
					{:else}
						<MoonIcon />
					{/if}
				</button>
			</li>
			<li>
				<button class="secondary outline" onclick={handleLogout}> Logout </button>
			</li>
		</ul>
	</nav>
{/if}

<main class="container">{@render children()}</main>

<style>
	nav {
		border-bottom: 1px solid var(--pico-muted-border-color);
		margin-bottom: 2rem;
	}

	button {
		padding: 0.25rem 0.75rem;
		margin-bottom: 0;
	}

	.theme-toggle {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		height: 2rem;
		padding: 0;
		border: none;
		border-radius: 50%;
		background: transparent;
		color: var(--pico-contrast);
	}
</style>
