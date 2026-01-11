<script lang="ts">
	// Use $props for Svelte 5
	let { transactions = $bindable([]) } = $props();

	const formatCurrency = (cents: number) => {
		return (cents / 100).toLocaleString('de-DE', {
			style: 'currency',
			currency: 'EUR'
		});
	};

	async function deleteTransaction(id: number) {
		if (!confirm('Are you sure you want to delete this transaction?')) return;

		// API Contract: DELETE /api/transactions?id={int}
		const res = await fetch(`/api/transactions?id=${id}`, {
			method: 'DELETE'
		});

		if (res.ok) {
			// Reactively remove the transaction from the list
			transactions = transactions.filter((t) => t.id !== id);
		} else {
			alert('Failed to delete transaction');
		}
	}
</script>

<article>
	<header><strong>Recent Transactions</strong></header>
	{#if transactions.length > 0}
		<figure>
			<table role="grid">
				<thead>
					<tr>
						<th>Date</th>
						<th>Description</th>
						<th class="number">Amount</th>
						<th></th>
					</tr>
				</thead>
				<tbody>
					{#each transactions as t}
						<tr>
							<td>{new Date(t.date).toLocaleDateString()}</td>
							<td>{t.description}</td>
							<td
								class="number"
								style="color: {t.type === 'EXPENSE'
									? 'var(--pico-del-color)'
									: 'var(--pico-ins-color)'}"
							>
								{t.type === 'EXPENSE' ? '-' : '+'}{formatCurrency(t.amountInCents)}
							</td>
							<td>
								<button
									class="outline contrast delete-btn"
									onclick={() => deleteTransaction(t.id)}
									title="Delete"
								>
									<span aria-hidden="true">×</span>
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</figure>
	{:else}
		<p>No transactions found. Start by adding one!</p>
	{/if}
</article>

<style>
	.number {
		text-align: right;
		font-variant-numeric: tabular-nums;
	}
	.delete-btn {
		padding: 0 0.5rem;
		margin-bottom: 0;
		line-height: 1;
		border-color: transparent;
	}
	.delete-btn:hover {
		background-color: var(--pico-del-color);
		color: white;
	}
</style>
