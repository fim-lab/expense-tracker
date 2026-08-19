<script lang="ts">
	import type { Wallet } from '$lib/types';

	let { data } = $props();

	const hasDepots = $derived(data.depots.length > 0);

	function walletName(walletId: number) {
		return data.wallets.find((w: Wallet) => w.id === walletId)?.name || 'Unknown Wallet';
	}
</script>

<article>
	<header><strong>Depots</strong></header>
	{#if hasDepots}
		{#each data.depots as depot (depot.id)}
			<div class="depot">
				<a href="/depots/{depot.id}">{depot.name}</a>
				<small>{walletName(depot.walletId)}</small>
			</div>
		{/each}
	{:else}
		<p>No depots yet.</p>
	{/if}
	<a href="/depots/add" role="button">Add Depot</a>
</article>

<style>
	.depot {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem;
		margin-bottom: 0.75rem;
		background: var(--pico-card-background-color);
		border-radius: var(--pico-border-radius);
		box-shadow: var(--pico-card-box-shadow);
		border-left: 4px solid var(--pico-primary);
	}

	.depot small {
		color: var(--pico-muted-color);
	}
</style>
