<script lang="ts">
	import { onMount } from 'svelte';
	import { goto, invalidateAll, preloadData } from '$app/navigation';
	import { page } from '$app/state';
	import BudgetCard from '$lib/components/BudgetCard.svelte';
	import BudgetGroupCard from '$lib/components/BudgetGroupCard.svelte';
	import TransactionCard from '$lib/components/TransactionCard.svelte';
	import TransactionSearchForm from '$lib/components/TransactionSearchForm.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import WealthGroupCard from '$lib/components/WealthGroupCard.svelte';
	import { formatCurrency } from '$lib/utils';
	import type { Budget, BudgetGroup, Depot, Wallet } from '$lib/types';

	const pageNr = $derived(page.data.page);
	const pageSize = $derived(page.data.pageSize);
	const totalPages = $derived(Math.ceil(page.data.total / pageSize));
	const hasWallets = $derived(page.data?.wallets?.length > 0);
	const hasDepots = $derived(page.data?.depots?.length > 0);
	const hasDebts = $derived((page.data?.debtTotal ?? 0) > 0);

	const budgetGroups: BudgetGroup[] = $derived(page.data?.budgetGroups ?? []);
	const visibleBudgets = $derived((page.data?.budgets ?? []).filter((b: Budget) => b.visible));
	const hasBudgets = $derived(visibleBudgets.length > 0);

	type BudgetListItem =
		| { kind: 'budget'; budget: Budget }
		| { kind: 'group'; group: BudgetGroup; budgets: Budget[] };

	const budgetListItems = $derived.by(() => {
		const items: BudgetListItem[] = [];
		const groupItemByGroupId: Record<
			number,
			{ kind: 'group'; group: BudgetGroup; budgets: Budget[] }
		> = {};
		for (const budget of visibleBudgets) {
			const group = budget.groupId
				? budgetGroups.find((g: BudgetGroup) => g.id === budget.groupId)
				: undefined;
			if (!group) {
				items.push({ kind: 'budget', budget });
				continue;
			}
			let entry = groupItemByGroupId[group.id];
			if (!entry) {
				entry = { kind: 'group', group, budgets: [] };
				groupItemByGroupId[group.id] = entry;
				items.push(entry);
			}
			entry.budgets.push(budget);
		}
		return items;
	});

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

	const preloadedUrls = new Set<string>();

	function preloadOnce(url: string) {
		if (preloadedUrls.has(url)) return;
		preloadedUrls.add(url);
		preloadData(url);
	}

	$effect(() => {
		if (pageNr > 1) preloadOnce(pageUrl(pageNr - 1));
		if (pageNr < totalPages) preloadOnce(pageUrl(pageNr + 1));
	});

	onMount(() => {
		if (hasDepots) {
			fetch('/api/portfolio/refresh-stale-prices', { method: 'POST' }).catch((err) => {
				console.error('Failed to refresh stale stock prices', err);
			});
		}
	});

	let widgetsOpen = $state(false);
	let filtersOpen = $state(false);

	function handleKeydown(e: KeyboardEvent) {
		const target = e.target as HTMLElement;
		const tag = target?.tagName;
		if (tag === 'INPUT' || tag === 'SELECT' || tag === 'TEXTAREA' || target?.isContentEditable)
			return;

		if (e.key === 'ArrowLeft' && pageNr > 1) {
			goto(pageUrl(pageNr - 1), { noScroll: true });
		} else if (e.key === 'ArrowRight' && pageNr < totalPages) {
			goto(pageUrl(pageNr + 1), { noScroll: true });
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="grid">
	<aside>
		{#if hasWallets || hasDepots || hasDebts}
			<WealthGroupCard
				totalWealthCents={totalWealth}
				wallets={page.data.wallets}
				depots={page.data.depots}
				{hasDebts}
				debtSumInCents={page.data.debtSumInCents}
				{walletName}
			/>
		{/if}
		{#if hasBudgets || hasWallets || hasDepots || hasDebts}
			<button
				type="button"
				class="mobile-toggle"
				aria-expanded={widgetsOpen}
				onclick={() => (widgetsOpen = !widgetsOpen)}
			>
				<span class="caret">{widgetsOpen ? '▾' : '▸'}</span> Budgets
			</button>
			<div class="collapsible" class:collapsed={!widgetsOpen}>
				<article>
					{#if hasBudgets}
						{#each budgetListItems as item (item.kind === 'group' ? `group-${item.group.id}` : `budget-${item.budget.id}`)}
							{#if item.kind === 'group'}
								<BudgetGroupCard group={item.group} budgets={item.budgets} />
							{:else}
								<BudgetCard budget={item.budget} />
							{/if}
						{/each}
					{/if}
				</article>
			</div>
		{/if}
	</aside>

	<aside>
		<button
			type="button"
			class="mobile-toggle"
			aria-expanded={filtersOpen}
			onclick={() => (filtersOpen = !filtersOpen)}
		>
			<span class="caret">{filtersOpen ? '▾' : '▸'}</span> Filters
		</button>
		<div class="collapsible" class:collapsed={!filtersOpen}>
			<article>
				<TransactionSearchForm budgets={page.data.budgets} wallets={page.data.wallets} />
			</article>
		</div>
	</aside>
	<article>
		{#if page.data.transactions?.length > 0}
			<header>
				<p class="total">
					Sum <strong>{formatCurrency(page.data.sumInCents)}</strong>
				</p>
			</header>{/if}

		<div class="transaction-list">
			{#if page.data.transactions?.length > 0}
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
		margin-bottom: 0;
		justify-content: space-between;
		color: var(--pico-muted-color);
	}

	.mobile-toggle {
		display: none;
	}

	@media (max-width: 768px) {
		.mobile-toggle {
			display: flex;
			align-items: center;
			gap: 0.25rem;
			width: auto;
			margin: 0 0 0.5rem;
			padding: 0;
			background: none;
			border: none;
			color: inherit;
			font-weight: 600;
			cursor: pointer;
		}

		.caret {
			color: var(--pico-muted-color);
		}

		.collapsible.collapsed {
			display: none;
		}
	}
</style>
