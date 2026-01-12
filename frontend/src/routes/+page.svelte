<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import TransactionCard from '$lib/components/TransactionCard.svelte';
	let { data } = $props();

	const pageCount = $derived(Math.ceil(data.total / data.limit));
	const pages = $derived(Array.from({ length: pageCount }, (_, i) => i + 1));

	async function deleteTransaction(id: number) {
		if (!confirm('Are you sure you want to delete this transaction?')) return;

		const res = await fetch(`/api/transactions?id=${id}`, {
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
			<ul>
				{#each data.wallets as wallet}
					<li>{wallet.name}</li>
				{/each}
			</ul>
		</article>
	</aside>

	<article>
		<header><strong>Recent Transactions</strong></header>

		<div class="transaction-list">
			{#each data.transactions as tx (tx.id)}
				<TransactionCard transaction={tx} ondelete={deleteTransaction} />
			{/each}
		</div>

		<nav aria-label="Pagination">
			<ul>
				{#each pages as p}
					<li>
						<a href={`?page=${p}`} aria-current={p === data.page ? 'page' : undefined}>
							{p}
						</a>
					</li>
				{/each}
			</ul>
		</nav>
	</article>
</div>

<style>
	.full-width {
		width: 100%;
	}
	.transaction-list {
		margin-bottom: 1rem;
	}
</style>
