<script lang="ts">
	import type { Wallet } from '$lib/types';

	let { data } = $props();

	let wallets = $state<Wallet[]>(
		// svelte-ignore state_referenced_locally
		data.wallets.map((w: Wallet) => ({ ...w, isEditing: false, newName: '' }))
	);

	function startEditing(wallet: Wallet) {
		wallet.isEditing = true;
		wallet.newName = wallet.name;
	}

	function cancelEditing(wallet: Wallet) {
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
</script>

<h1>Wallets</h1>

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
							<button class="secondary" onclick={() => cancelEditing(wallet)}>Cancel</button>
						{:else}
							<button onclick={() => startEditing(wallet)}>Edit</button>
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
	input {
		margin-bottom: 0;
	}
</style>
