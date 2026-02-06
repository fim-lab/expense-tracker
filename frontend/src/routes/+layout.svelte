<script lang="ts">
	import { page } from '$app/state';
	import { browser } from '$app/environment';

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
		<ul><li><strong>FinanceTracker</strong></li></ul>
		<ul>
			<li><a href="/">Dashboard</a></li>
			<li><a href="/transactions/add">Add TX</a></li>
			<li><a href="/salary">Salary</a></li>
			<li><a href="/budgets/add">Add Budget</a></li>
			<li><a href="/wallets/add">Add Wallet</a></li>
			<li><a href="/wallets/transfer">Transfer</a></li>
			<li><a href="/settings">Settings</a></li>
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
</style>
