<script lang="ts">
	import { page } from '$app/state'; // The new reactive state
	import { browser } from '$app/environment';

	// Accessing page.url.pathname is now natively reactive in Svelte 5
	// No $derived needed unless you want to store it in a local variable

	async function handleLogout() {
		const res = await fetch('/auth/logout', { method: 'POST' });
		if (res.ok && browser) {
			window.location.href = '/login';
		}
	}

	let { children } = $props();
</script>

{#if page.url.pathname !== '/login'}
	<nav class="container">
		<ul>
			<li><strong>FinanceTracker</strong></li>
		</ul>
		<ul>
			<li><a href="/">Dashboard</a></li>
			<li><a href="/transactions/add">Add TX</a></li>
			<li><a href="/budgets/add">Add Budget</a></li>
			<li><a href="/wallets/add">Add Wallet</a></li>
			<li>
				<button class="outline secondary" onclick={handleLogout}>Logout</button>
			</li>
		</ul>
	</nav>
{/if}

<main class="container">
	{@render children()}
</main>

<style>
	nav {
		border-bottom: 1px solid var(--pico-muted-border-color);
		margin-bottom: 2rem;
	}
	button {
		padding: 0.25rem 0.75rem;
		margin-bottom: 0;
	}
</style>
