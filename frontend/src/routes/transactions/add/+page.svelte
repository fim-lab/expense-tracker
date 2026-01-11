<script lang="ts">
	import { goto } from '$app/navigation';
	let { data } = $props();

	let description = $state('');
	let amountInCents = $state(0);
	let date = $state(new Date().toISOString().split('T')[0]);
	let walletId = $state(0);
	let budgetId = $state(0);
	let type = $state('EXPENSE');

	async function handleSubmit(e: Event) {
		e.preventDefault();

		// 1. Convert YYYY-MM-DD to RFC3339 for Go's time.Time
		const rfc3339Date = new Date(date).toISOString();

		const payload = {
			// ID is 0 for new records (Go ignores it or uses it as 'unset')
			id: 0,
			// UserID is usually handled by the backend session,
			// but if the struct requires it, we send 0
			userId: 0,
			date: rfc3339Date,
			description: description,
			amountInCents: Math.round(amountInCents * 100),
			walletId: Number(walletId),
			budgetId: Number(budgetId),
			type: type,
			isPending: false, // Explicitly send boolean
			tags: [] // Send empty array to match []string
		};

		const res = await fetch('/api/transactions', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});

		if (res.ok) {
			goto('/');
		} else {
			const errorText = await res.text();
			console.error('Backend Error:', errorText);
			alert('Failed to save: ' + errorText);
		}
	}
</script>

<article>
	<h3>Add New Transaction</h3>
	<form onsubmit={handleSubmit}>
		<div class="grid">
			<label>
				Date
				<input type="date" bind:value={date} required />
			</label>
			<label>
				Type
				<select bind:value={type}>
					<option value="EXPENSE">Expense</option>
					<option value="INCOME">Income</option>
				</select>
			</label>
		</div>

		<label>
			Description
			<input type="text" bind:value={description} placeholder="Grocery shopping..." required />
		</label>

		<div class="grid">
			<label>
				Amount (EUR)
				<input type="number" step="0.01" bind:value={amountInCents} required />
			</label>
			<label>
				Wallet
				<select bind:value={walletId} required>
					<option value={0} disabled>Select Wallet</option>
					{#each data.wallets as wallet}
						<option value={wallet.id}>{wallet.name}</option>
					{/each}
				</select>
			</label>
		</div>

		<label>
			Budget
			<select bind:value={budgetId} required>
				<option value={0} disabled>Select Budget Category</option>
				{#each data.budgets as budget}
					<option value={budget.id}>{budget.name}</option>
				{/each}
			</select>
		</label>

		<button type="submit">Save Transaction</button>
	</form>
</article>
