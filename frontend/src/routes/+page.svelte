<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { page } from '$app/state';
	import BudgetCard from '$lib/components/BudgetCard.svelte';
	import DebtCard from '$lib/components/DebtCard.svelte';
	import DepotCard from '$lib/components/DepotCard.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import TransactionCard from '$lib/components/TransactionCard.svelte';
	import TransactionSearchForm from '$lib/components/TransactionSearchForm.svelte';
	import WalletCard from '$lib/components/WalletCard.svelte';
	import { formatCurrency } from '$lib/utils';
	import type { Depot, Wallet } from '$lib/types';

	const pageNr = page.data.page;
	const pageSize = page.data.pageSize;
	const totalPages = $derived(Math.ceil(page.data.total / pageSize));
	const hasWallets = $derived(page.data?.wallets?.length > 0);
	const hasDepots = $derived(page.data?.depots?.length > 0);
	const hasBudgets = $derived(page.data?.budgets?.length > 0);
	const hasDebts = $derived((page.data?.debtTotal ?? 0) > 0);
	const totalWealth = $derived(
		(page.data?.wallets ?? []).reduce((sum: number, w: Wallet) => sum + w.balanceCents, 0) +
			(page.data?.depots ?? []).reduce(
				(sum: number, d: Depot) => sum + (d.currentValueInCents ?? d.investedInCents ?? 0),
				0
			) +
			(page.data?.debtSumInCents ?? 0)
	);

	async function deleteTransaction(id: number) {
		if (!confirm('Are you sure you want to delete this transaction?')) return;

		const res = await fetch(`/api/transactions/${id}`, {
			method: 'DELETE'
		});

		if (res && res.ok) {
			await invalidateAll();
		} else {
			alert('Failed to delete transaction');
		}
	}

	async function importTestData() {
		const res = await fetch('/api/transactions/import/testdata', {
			method: 'POST'
		});

		if (res && res.ok) {
			await invalidateAll();
		} else {
			alert('Failed to import test data');
		}
	}

	const isSearchActive = $derived(page.url.searchParams.size > 0);

	function walletName(walletId: number) {
		return page.data.wallets?.find((w: Wallet) => w.id === walletId)?.name ?? '';
	}
</script>

<div class="grid">
	<aside>
		<article>
			<header><strong>Search</strong></header>
			<TransactionSearchForm budgets={page.data.budgets} wallets={page.data.wallets} />
		</article>
		{#if hasBudgets || hasWallets || hasDepots || hasDebts}
			<article>
				{#if hasWallets || hasDepots || hasDebts}
					<p class="total">
						Total wealth <strong>{formatCurrency(totalWealth)}</strong>
					</p>
				{/if}
				{#if hasWallets}
					<header><strong>Wallets</strong></header>
					{#each page.data.wallets as wallet}
						<WalletCard {wallet} />
					{/each}
				{/if}
				{#if hasDebts}
					<header><strong>Debt</strong></header>
					<DebtCard amountInCents={page.data.debtSumInCents} />
				{/if}
				{#if hasDepots}
					<header><strong>Depots</strong></header>
					{#each page.data.depots as depot (depot.id)}
						<DepotCard {depot} subtitle={walletName(depot.walletId)} />
					{/each}
				{/if}
				{#if hasBudgets}
					<header><strong>Budgets</strong></header>
					{#each page.data.budgets as budget}
						<BudgetCard {budget} />
					{/each}
				{/if}
			</article>
		{/if}
	</aside>

	<article>
		<header><strong>Recent Transactions</strong></header>

		<div class="transaction-list">
			{#if page.data.transactions?.length > 0}
				<p class="total">
					Sum <strong>{formatCurrency(page.data.sumInCents)}</strong>
				</p>
				{#each page.data.transactions as tx (tx.id)}
					<TransactionCard transaction={tx} ondelete={deleteTransaction} />
				{/each}
			{:else}
				<p>No transactions found.</p>
				{#if page.data.total === 0 && !isSearchActive}
					<button on:click={importTestData}>Import Test Data</button>
				{/if}
			{/if}
		</div>
		{#if totalPages > 1}
			<Pagination page={pageNr} {totalPages} />
		{/if}
	</article>
</div>

<style>
	.total {
		display: flex;
		justify-content: space-between;
		color: var(--pico-muted-color);
	}
</style>
