<script lang="ts">
	import DepotCard from '$lib/components/DepotCard.svelte';
	import { formatCurrency } from '$lib/utils';
	import type { Depot, Wallet, Budget } from '$lib/types';

	let { data } = $props();

	const hasDepots = $derived(data.depots.length > 0);
	const totalInvested = $derived(
		data.depots.reduce((sum: number, depot: Depot) => sum + (depot.investedInCents ?? 0), 0)
	);

	function walletName(walletId: number) {
		return data.wallets.find((w: Wallet) => w.id === walletId)?.name || 'Unknown Wallet';
	}

	function budgetName(budgetId: number) {
		return data.budgets.find((b: Budget) => b.id === budgetId)?.name || 'Unknown Budget';
	}
</script>

<article>
	<header><strong>Depots</strong></header>
	{#if hasDepots}
		{#each data.depots as depot (depot.id)}
			<DepotCard
				{depot}
				href="/depots/{depot.id}"
				subtitle={`${walletName(depot.walletId)} · ${budgetName(depot.budgetId)}`}
			/>
		{/each}
		<p class="total">
			Invested in total <strong>{formatCurrency(totalInvested)}</strong>
		</p>
	{:else}
		<p>No depots yet.</p>
	{/if}
	<a href="/depots/add" role="button">Add Depot</a>
</article>

<style>
	.total {
		display: flex;
		justify-content: space-between;
		color: var(--pico-muted-color);
	}
</style>
