<script lang="ts">
	import { goto, invalidateAll, preloadData } from '$app/navigation';
	import { page } from '$app/state';
	import BudgetCard from '$lib/components/BudgetCard.svelte';
	import DebtCard from '$lib/components/DebtCard.svelte';
	import DepotCard from '$lib/components/DepotCard.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import TransactionCard from '$lib/components/TransactionCard.svelte';
	import TransactionSearchForm from '$lib/components/TransactionSearchForm.svelte';
	import WalletCard from '$lib/components/WalletCard.svelte';
	import { formatCurrency, updateParams } from '$lib/utils';
	import type { Budget, BudgetGroup, Depot, Wallet } from '$lib/types';

	const pageNr = $derived(page.data.page);
	const pageSize = $derived(page.data.pageSize);
	const totalPages = $derived(Math.ceil(page.data.total / pageSize));
	const hasWallets = $derived(page.data?.wallets?.length > 0);
	const hasDepots = $derived(page.data?.depots?.length > 0);
	const hasDebts = $derived((page.data?.debtTotal ?? 0) > 0);

	const budgetGroups: BudgetGroup[] = $derived(page.data?.budgetGroups ?? []);
	const activeGroupId = $derived.by(() => {
		const id = page.url.searchParams.get('budget_group_id');
		return id ? Number(id) : undefined;
	});
	const visibleBudgets = $derived(
		(activeGroupId
			? (page.data?.budgets ?? []).filter((b: Budget) => b.groupId === activeGroupId)
			: (page.data?.budgets ?? [])
		).filter((b: Budget) => !b.isDormant)
	);
	const hasBudgets = $derived(visibleBudgets.length > 0);

	function selectGroup(groupId: number | undefined) {
		updateParams({ budget_group_id: groupId });
	}
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

	const pageUrl = (p: number) => {
		const url = new URL(page.url);
		url.searchParams.set('page', p.toString());
		return url.toString();
	};

	$effect(() => {
		if (pageNr > 1) preloadData(pageUrl(pageNr - 1));
		if (pageNr < totalPages) preloadData(pageUrl(pageNr + 1));
	});

	function handleKeydown(e: KeyboardEvent) {
		const target = e.target as HTMLElement;
		const tag = target?.tagName;
		if (tag === 'INPUT' || tag === 'SELECT' || tag === 'TEXTAREA' || target?.isContentEditable)
			return;

		if (e.key === 'ArrowLeft' && pageNr > 1) {
			goto(pageUrl(pageNr - 1));
		} else if (e.key === 'ArrowRight' && pageNr < totalPages) {
			goto(pageUrl(pageNr + 1));
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="grid">
	<aside>
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
					{#each visibleBudgets as budget}
						<BudgetCard {budget} />
					{/each}
				{/if}
			</article>
		{/if}
	</aside>

	<aside>
		<article>
			<header><strong>Search</strong></header>
			<TransactionSearchForm budgets={page.data.budgets} wallets={page.data.wallets} />
		</article>
	</aside>
	<article>
		<header><strong>Recent Transactions</strong></header>

		{#if budgetGroups.length > 0}
			<div class="group-tabs">
				<button
					type="button"
					class="secondary outline"
					class:active={activeGroupId === undefined}
					onclick={() => selectGroup(undefined)}
				>
					All
				</button>
				{#each budgetGroups as group (group.id)}
					<button
						type="button"
						class="secondary outline"
						class:active={activeGroupId === group.id}
						onclick={() => selectGroup(group.id)}
					>
						{group.name}
					</button>
				{/each}
			</div>
		{/if}

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
					<button onclick={importTestData}>Import Test Data</button>
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

	.group-tabs {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin-bottom: 1rem;
	}

	.group-tabs button {
		width: auto;
		margin: 0;
		padding: 0.25rem 0.75rem;
		font-size: 0.85rem;
	}

	.group-tabs button.active {
		background-color: var(--pico-primary);
		color: var(--pico-primary-inverse);
	}
</style>
