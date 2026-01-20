<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import BudgetCard from '$lib/components/BudgetCard.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import TransactionCard from '$lib/components/TransactionCard.svelte';
	import TransactionSearchForm from '$lib/components/TransactionSearchForm.svelte';
	import WalletCard from '$lib/components/WalletCard.svelte';
	import { updateParams } from '$lib/utils';

	const pageNr = page.data.page
	const pageSize = page.data.pageSize
	const totalPages = $derived(Math.ceil(page.data.total / pageSize));

	async function deleteTransaction(id: number) {
		if (!confirm('Are you sure you want to delete this transaction?')) return;

		const res = await fetch(`/api/transactions/${id}`, {
			method: 'DELETE'
		});

		if (res && res.ok) {
			await updateParams({});
		} else {
			alert('Failed to delete transaction');
		}
	}
</script>

<div class="grid">
	<aside>
		<article>
			<header><strong>Search</strong></header>
			<TransactionSearchForm budgets={page.data.budgets} wallets={page.data.wallets} />
		</article>
		<article>
			<header><strong>Wallets</strong></header>
			{#each page.data.wallets as wallet}
				<WalletCard {wallet} />
			{/each}
			<header><strong>Budgets</strong></header>
			{#each page.data.budgets as budget}
				<BudgetCard {budget} />
			{/each}
		</article>
	</aside>

	<article>
		<header><strong>Recent Transactions</strong></header>

		<div class="transaction-list">
			{#if page.data.transactions?.length > 0}
				{#each page.data.transactions as tx (tx.id)}
					<TransactionCard transaction={tx} ondelete={deleteTransaction} />
				{/each}
			{:else}
				<p>No transactions found.</p>
			{/if}
		</div>
		{#if totalPages > 1}
			<Pagination page={pageNr} {totalPages} />
		{/if}
	</article>
</div>
