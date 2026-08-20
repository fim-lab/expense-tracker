<script lang="ts">
	import type { Wallet, Depot, Budget } from '$lib/types';

	let { data } = $props();

	let wallets = $state<Wallet[]>(
		// svelte-ignore state_referenced_locally
		(data.wallets || []).map((w: Wallet) => ({ ...w, isEditing: false, newName: '' }))
	);

	let budgets = $state<Budget[]>(
		// svelte-ignore state_referenced_locally
		data.budgets || []
	);

	let depots = $state<Depot[]>(
		// svelte-ignore state_referenced_locally
		data.depots?.map((d: Depot) => ({
			...d,
			isEditing: false,
			newName: '',
			newWalletId: d.walletId,
			newBudgetId: d.budgetId
		}))
	);

	// Wallets logic
	function startEditingWallet(wallet: Wallet) {
		wallet.isEditing = true;
		wallet.newName = wallet.name;
	}

	function cancelEditingWallet(wallet: Wallet) {
		wallet.isEditing = false;
	}

	async function updateWallet(wallet: Wallet) {
		if (!wallet.newName) {
			alert('Enter a name.');
			return;
		}
		const res = await fetch(`/api/wallets/${wallet.id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ ...wallet, name: wallet.newName })
		});

		if (res.ok) {
			wallet.name = wallet.newName;
			wallet.isEditing = false;
		} else {
			console.error('Failed to update wallet');
		}
	}

	async function deleteWallet(walletId: number) {
		if (confirm('Are you sure you want to delete this wallet?')) {
			const res = await fetch(`/api/wallets/${walletId}`, {
				method: 'DELETE'
			});

			if (res.ok) {
				wallets = wallets.filter((w) => w.id !== walletId);
			} else {
				console.error('Failed to delete wallet');
			}
		}
	}

	// Depots logic
	function startEditingDepot(depot: Depot) {
		depot.isEditing = true;
		depot.newName = depot.name;
		depot.newWalletId = depot.walletId;
		depot.newBudgetId = depot.budgetId;
	}

	function cancelEditingDepot(depot: Depot) {
		depot.isEditing = false;
	}

	async function updateDepot(depot: Depot) {
		if (!depot.newName) {
			alert('Enter a name.');
			return;
		}
		const res = await fetch(`/api/depots/${depot.id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				...depot,
				name: depot.newName,
				walletId: depot.newWalletId,
				budgetId: depot.newBudgetId
			})
		});

		if (res.ok) {
			depot.name = depot.newName;
			depot.walletId = depot.newWalletId!;
			depot.budgetId = depot.newBudgetId!;
			depot.isEditing = false;
		} else {
			console.error('Failed to update depot');
			alert('Failed to update depot');
		}
	}

	async function deleteDepot(depotId: number) {
		if (confirm('Are you sure you want to delete this depot?')) {
			const res = await fetch(`/api/depots/${depotId}`, {
				method: 'DELETE'
			});

			if (res.ok) {
				depots = depots.filter((d) => d.id !== depotId);
			} else {
				// A depot that still holds trades cannot be deleted; show why.
				const reason = await res.text();
				console.error('Failed to delete depot', reason);
				alert(reason || 'Failed to delete depot');
			}
		}
	}

	function getWalletName(walletId: number) {
		return wallets.find((w) => w.id === walletId)?.name || 'Unknown Wallet';
	}

	function getBudgetName(budgetId: number) {
		return budgets.find((b) => b.id === budgetId)?.name || 'Unknown Budget';
	}
</script>

<h1>Wallets & Depots</h1>

<section>
	<h2>Wallets</h2>
	{#if wallets.length > 0}
		<table>
			<thead>
				<tr>
					<th>Name</th>
					<th>Actions</th>
				</tr>
			</thead>
			<tbody>
				{#each wallets as wallet (wallet.id)}
					<tr>
						<td>
							{#if wallet.isEditing}
								<input type="text" bind:value={wallet.newName} />
							{:else}
								{wallet.name}
							{/if}
						</td>
						<td>
							{#if wallet.isEditing}
								<button onclick={() => updateWallet(wallet)}>OK</button>
								<button class="secondary" onclick={() => cancelEditingWallet(wallet)}>Cancel</button
								>
							{:else}
								<button onclick={() => startEditingWallet(wallet)}>Edit</button>
								<span
									title={!wallet.canDelete
										? 'Only budgets with a balance of 0 and no transactions can be deleted.'
										: ''}
								>
									<button
										class="secondary"
										onclick={() => deleteWallet(wallet.id)}
										disabled={!wallet.canDelete}>Delete</button
									>
								</span>
							{/if}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{:else}
		<p>No wallets found.</p>
	{/if}
	<a href="/wallets/add" role="button">Add Wallet</a>
</section>

<hr />

<section>
	<h2>Depots</h2>
	{#if depots?.length > 0}
		<table>
			<thead>
				<tr>
					<th>Name</th>
					<th>Wallet</th>
					<th>Budget</th>
					<th>Actions</th>
				</tr>
			</thead>
			<tbody>
				{#each depots as depot (depot.id)}
					<tr>
						<td>
							{#if depot.isEditing}
								<input type="text" bind:value={depot.newName} />
							{:else}
								<a href="/depots/{depot.id}">{depot.name}</a>
							{/if}
						</td>
						<td>
							{#if depot.isEditing}
								<select bind:value={depot.newWalletId}>
									{#each wallets as wallet}
										<option value={wallet.id}>{wallet.name}</option>
									{/each}
								</select>
							{:else}
								{getWalletName(depot.walletId)}
							{/if}
						</td>
						<td>
							{#if depot.isEditing}
								<select bind:value={depot.newBudgetId}>
									{#each budgets as budget}
										<option value={budget.id}>{budget.name}</option>
									{/each}
								</select>
							{:else}
								{getBudgetName(depot.budgetId)}
							{/if}
						</td>
						<td>
							{#if depot.isEditing}
								<button onclick={() => updateDepot(depot)}>OK</button>
								<button class="secondary" onclick={() => cancelEditingDepot(depot)}>Cancel</button>
							{:else}
								<button onclick={() => startEditingDepot(depot)}>Edit</button>
								<button class="secondary" onclick={() => deleteDepot(depot.id)}>Delete</button>
							{/if}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{:else}
		<p>No depots found.</p>
	{/if}
	<a href="/depots/add" role="button">Add Depot</a>
</section>

<style>
	table {
		width: 100%;
	}
	th,
	td {
		text-align: left;
		padding: 0.5rem;
	}
	th:last-child,
	td:last-child {
		text-align: right;
		white-space: nowrap;
	}
	input,
	select {
		margin-bottom: 0;
	}
	section {
		margin-bottom: 2rem;
	}
</style>
