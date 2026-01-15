<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { page } from '$app/state';
	import TransactionCard from '$lib/components/TransactionCard.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import BudgetCard from '$lib/components/BudgetCard.svelte';
	import WalletCard from '$lib/components/WalletCard.svelte';

	const totalPages = $derived(Math.ceil(page.data.total / page.data.limit));

	async function deleteTransaction(id: number) {
		if (!confirm('Are you sure you want to delete this transaction?')) return;

		const res = await fetch(`/api/transactions/${id}`, {
			method: 'DELETE'
		});

		if (res.ok) {
			await invalidateAll();
		} else {
			alert('Failed to delete transaction');
		}
	}
</script>

<div class="grid">
	<aside>
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
			{#each page.data.transactions as tx (tx.id)}
				<TransactionCard transaction={tx} ondelete={deleteTransaction} />
			{/each}
		</div>
		{#if totalPages > 1}
			<Pagination page={page.data.page} {totalPages} />
		{/if}
	</article>
</div>
