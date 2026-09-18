<script lang="ts">
	import DebtCard from './DebtCard.svelte';
	import DepotCard from './DepotCard.svelte';
	import WalletCard from './WalletCard.svelte';
	import { isDesktopViewport } from '$lib/viewport';
	import { formatCurrency } from '$lib/utils';
	import type { Depot, Wallet } from '$lib/types';

	let {
		totalWealthCents,
		wallets,
		depots,
		hasDebts,
		debtSumInCents,
		walletName
	}: {
		totalWealthCents: number;
		wallets: Wallet[];
		depots: Depot[];
		hasDebts: boolean;
		debtSumInCents: number;
		walletName: (walletId: number) => string;
	} = $props();

	let expanded = $state(isDesktopViewport());
</script>

<button
	type="button"
	class="card wealth-card"
	aria-expanded={expanded}
	onclick={() => (expanded = !expanded)}
>
	<p class="card-title">
		<span><span class="caret">{expanded ? '▾' : '▸'}</span> Total wealth</span>
		<span class="card-amount">{formatCurrency(totalWealthCents)}</span>
	</p>
</button>
{#if expanded}
	<div class="wealth-members">
		{#each wallets as wallet (wallet.id)}
			<WalletCard {wallet} />
		{/each}
		{#each depots as depot (depot.id)}
			<DepotCard {depot} href="/depots/{depot.id}" subtitle={walletName(depot.walletId)} />
		{/each}
		{#if hasDebts}
			<DebtCard amountInCents={debtSumInCents} />
		{/if}
	</div>
{/if}

<style>
	.card {
		display: flex;
		justify-content: space-between;
		align-items: center;
		width: 100%;
		padding: 1rem;
		margin-bottom: 0.75rem;
		background: var(--pico-card-background-color);
		border-radius: var(--pico-border-radius);
		box-shadow: var(--pico-card-box-shadow);
		border-left: 4px solid var(--pico-primary);
		color: inherit;
		text-align: left;
		cursor: pointer;

		p {
			color: inherit;
		}
	}

	.wealth-card {
		border: none;
		border-left: 4px solid var(--pico-primary);
		font: inherit;
	}

	.caret {
		color: var(--pico-muted-color);
	}

	.card-title {
		font-weight: bold;
		margin-bottom: 0;
		display: flex;
		justify-content: space-between;
		width: 100%;
	}

	.card-amount {
		font-weight: normal;
		font-size: 0.9em;
		font-family: ui-monospace, SFMono-Regular, monospace;
	}

	.wealth-members {
		margin-left: 1rem;
		padding-left: 0.75rem;
		border-left: 2px solid var(--pico-muted-border-color);
	}
</style>
