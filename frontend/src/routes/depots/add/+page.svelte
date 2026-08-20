<script lang="ts">
	import { goto } from '$app/navigation';

	let { data } = $props();

	let name = $state('');
	let walletId = $state<number>(data.wallets?.[0]?.id || 0);
	let budgetId = $state<number>(data.budgets?.[0]?.id || 0);
	let isSubmitting = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (walletId === 0 || budgetId === 0) {
			alert('Please select a wallet and a budget.');
			return;
		}
		isSubmitting = true;

		const payload = {
			id: 0,
			userId: 0,
			name: name,
			walletId: walletId,
			budgetId: budgetId
		};

		const res = await fetch('/api/depots', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});

		if (res.ok) {
			goto('/settings/wallets');
		} else {
			alert('Failed to save depot');
			isSubmitting = false;
		}
	}
</script>

<article>
	<header><strong>Add New Depot</strong></header>
	<form onsubmit={handleSubmit}>
		<label>
			Depot Name
			<input type="text" bind:value={name} placeholder="e.g. My Depot" required />
		</label>

		<label>
			Associated Wallet
			<select bind:value={walletId} required>
				<option value={0} disabled>Select a wallet</option>
				{#each data.wallets || [] as wallet}
					<option value={wallet.id}>{wallet.name}</option>
				{/each}
			</select>
		</label>

		<label>
			Associated Budget
			<select bind:value={budgetId} required>
				<option value={0} disabled>Select a budget</option>
				{#each data.budgets || [] as budget}
					<option value={budget.id}>{budget.name}</option>
				{/each}
			</select>
		</label>

		<button type="submit" aria-busy={isSubmitting}>Create Depot</button>
	</form>
</article>
