<script lang="ts">
	import type { Depot, Wallet } from '$lib/types';

	let { data } = $props();

	let depots = $state<Depot[]>(
		// svelte-ignore state_referenced_locally
		data.depots?.map((d: Depot) => ({
			...d,
			isEditing: false,
			newName: '',
			newWalletId: d.walletId
		}))
	);
	let wallets = $state<Wallet[]>(data.wallets);

	function startEditing(depot: Depot) {
		depot.isEditing = true;
		depot.newName = depot.name;
		depot.newWalletId = depot.walletId;
	}

	function cancelEditing(depot: Depot) {
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
			body: JSON.stringify({ ...depot, name: depot.newName, walletId: depot.newWalletId })
		});

		if (res.ok) {
			depot.name = depot.newName;
			depot.walletId = depot.newWalletId!;
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
				console.error('Failed to delete depot');
				alert('Failed to delete depot');
			}
		}
	}

	function getWalletName(walletId: number) {
		return wallets.find((w) => w.id === walletId)?.name || 'Unknown Wallet';
	}
</script>

<h1>Depots</h1>

{#if depots?.length > 0}
	<table>
		<thead>
			<tr>
				<th>Name</th>
				<th>Wallet</th>
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
							{depot.name}
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
							<button onclick={() => updateDepot(depot)}>OK</button>
							<button class="secondary" onclick={() => cancelEditing(depot)}>Cancel</button>
						{:else}
							<button onclick={() => startEditing(depot)}>Edit</button>
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

<a href="/depots/add">Add Depot</a>

<style>
	table {
		width: 100%;
	}
	th,
	td {
		text-align: left;
		padding: 0.5rem;
	}
	td:last-child {
		text-align: right;
		white-space: nowrap;
	}
	input,
	select {
		margin-bottom: 0;
	}
</style>
