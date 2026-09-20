<script lang="ts">
	import type { Wallet, Depot, Budget, Stock } from '$lib/types';
	import SaveIcon from '$lib/components/icons/SaveIcon.svelte';
	import DeleteIcon from '$lib/components/icons/DeleteIcon.svelte';
	import RefreshIcon from '$lib/components/icons/RefreshIcon.svelte';

	let { data } = $props();

	let wallets = $state<Wallet[]>(
		// svelte-ignore state_referenced_locally
		(data.wallets || []).map((w: Wallet) => ({ ...w, newName: w.name }))
	);

	let budgets = $state<Budget[]>(
		// svelte-ignore state_referenced_locally
		data.budgets || []
	);

	let depots = $state<Depot[]>(
		// svelte-ignore state_referenced_locally
		data.depots?.map((d: Depot) => ({
			...d,
			newName: d.name,
			newWalletId: d.walletId,
			newBudgetId: d.budgetId
		}))
	);

	// Wallets logic
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

	let stocks = $state<Stock[]>(
		// svelte-ignore state_referenced_locally
		(data.stocks || []).map((s: Stock) => ({
			...s,
			newWkn: s.wkn,
			newTicker: s.ticker,
			newPriceEuros: s.priceInCents / 100
		}))
	);

	let newStockWkn = $state('');
	let newStockTicker = $state('');
	let newStockPriceEuros = $state<number | undefined>(undefined);
	let refreshingStockId = $state<number | null>(null);

	function formatLastFetched(lastFetched: string | null) {
		if (!lastFetched) return 'never';
		const date = new Date(lastFetched);
		if (Number.isNaN(date.getTime()) || date.getFullYear() <= 1) return 'never';
		return date.toLocaleDateString();
	}

	async function updateStock(stock: Stock) {
		if (!stock.newWkn) {
			alert('Enter a WKN.');
			return;
		}
		const res = await fetch(`/api/stocks/${stock.id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				...stock,
				wkn: stock.newWkn,
				ticker: stock.newTicker ?? '',
				priceInCents: Math.round((stock.newPriceEuros ?? 0) * 100)
			})
		});

		if (res.ok) {
			const updated = await res.json();
			stock.wkn = updated.wkn;
			stock.ticker = updated.ticker;
			stock.priceInCents = updated.priceInCents;
			stock.lastFetched = updated.lastFetched;
		} else {
			console.error('Failed to update stock');
			alert('Failed to update stock');
		}
	}

	async function refreshStockPrice(stock: Stock, confirmPriceInCents?: number) {
		refreshingStockId = stock.id;
		try {
			const res = await fetch(`/api/stocks/${stock.id}/refresh`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(confirmPriceInCents !== undefined ? { confirmPriceInCents } : {})
			});

			if (!res.ok) {
				const reason = await res.text();
				alert(reason || 'Could not fetch the current price for this stock.');
				return;
			}

			const result = await res.json();
			if (result.needsConfirmation) {
				const oldEuros = (result.oldPriceInCents / 100).toFixed(2);
				const newEuros = (result.newPriceInCents / 100).toFixed(2);
				if (
					confirm(
						`The fetched price (€${newEuros}) differs by more than 10% from the current price (€${oldEuros}). Update anyway?`
					)
				) {
					await refreshStockPrice(stock, result.newPriceInCents);
				}
				return;
			}

			stock.priceInCents = result.stock.priceInCents;
			stock.lastFetched = result.stock.lastFetched;
			stock.newPriceEuros = result.stock.priceInCents / 100;
		} catch (err) {
			console.error('Failed to refresh stock price', err);
			alert('Could not fetch the current price for this stock.');
		} finally {
			refreshingStockId = null;
		}
	}

	async function deleteStock(stockId: number) {
		if (confirm('Are you sure you want to delete this stock?')) {
			const res = await fetch(`/api/stocks/${stockId}`, {
				method: 'DELETE'
			});

			if (res.ok) {
				stocks = stocks.filter((s) => s.id !== stockId);
			} else {
				const reason = await res.text();
				console.error('Failed to delete stock', reason);
				alert(reason || 'Failed to delete stock');
			}
		}
	}

	async function createStock() {
		if (!newStockWkn) {
			alert('Enter a WKN.');
			return;
		}
		const res = await fetch('/api/stocks', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				wkn: newStockWkn,
				ticker: newStockTicker,
				priceInCents: Math.round((newStockPriceEuros ?? 0) * 100)
			})
		});

		if (res.ok) {
			const created = await res.json();
			stocks = [
				...stocks,
				{
					...created,
					newWkn: created.wkn,
					newTicker: created.ticker,
					newPriceEuros: created.priceInCents / 100
				}
			];
			newStockWkn = '';
			newStockTicker = '';
			newStockPriceEuros = undefined;
		} else {
			console.error('Failed to create stock');
			alert('Failed to create stock');
		}
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
							<input type="text" bind:value={wallet.newName} />
						</td>
						<td>
							<button
								type="button"
								class="icon-button save-button"
								onclick={() => updateWallet(wallet)}
								aria-label="Save wallet"
								title="Save wallet"
							>
								<SaveIcon />
							</button>
							<span
								title={!wallet.canDelete
									? 'Only budgets with a balance of 0 and no transactions can be deleted.'
									: ''}
							>
								<button
									type="button"
									class="icon-button delete-button"
									onclick={() => deleteWallet(wallet.id)}
									disabled={!wallet.canDelete}
									aria-label="Delete wallet"
									title="Delete wallet"
								>
									<DeleteIcon />
								</button>
							</span>
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
							<input type="text" bind:value={depot.newName} />
						</td>
						<td>
							<select bind:value={depot.newWalletId}>
								{#each wallets as wallet}
									<option value={wallet.id}>{wallet.name}</option>
								{/each}
							</select>
						</td>
						<td>
							<select bind:value={depot.newBudgetId}>
								{#each budgets as budget}
									<option value={budget.id}>{budget.name}</option>
								{/each}
							</select>
						</td>
						<td>
							<button
								type="button"
								class="icon-button save-button"
								onclick={() => updateDepot(depot)}
								aria-label="Save depot"
								title="Save depot"
							>
								<SaveIcon />
							</button>
							<button
								type="button"
								class="icon-button delete-button"
								onclick={() => deleteDepot(depot.id)}
								aria-label="Delete depot"
								title="Delete depot"
							>
								<DeleteIcon />
							</button>
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

<hr />

<section>
	<h2>Stocks</h2>
	{#if stocks.length > 0}
		<table>
			<thead>
				<tr>
					<th>WKN</th>
					<th>Ticker</th>
					<th>Price</th>
					<th>Last Fetched</th>
					<th>Actions</th>
				</tr>
			</thead>
			<tbody>
				{#each stocks as stock (stock.id)}
					<tr>
						<td>
							<input type="text" bind:value={stock.newWkn} />
						</td>
						<td>
							<input type="text" bind:value={stock.newTicker} />
						</td>
						<td>
							<input type="number" step="0.01" bind:value={stock.newPriceEuros} />
						</td>
						<td>{formatLastFetched(stock.lastFetched)}</td>
						<td>
							<button
								type="button"
								class="icon-button refresh-button"
								onclick={() => refreshStockPrice(stock)}
								disabled={refreshingStockId === stock.id}
								aria-label="Refresh price"
								title="Refresh price"
							>
								<RefreshIcon />
							</button>
							<button
								type="button"
								class="icon-button save-button"
								onclick={() => updateStock(stock)}
								aria-label="Save stock"
								title="Save stock"
							>
								<SaveIcon />
							</button>
							<button
								type="button"
								class="icon-button delete-button"
								onclick={() => deleteStock(stock.id)}
								aria-label="Delete stock"
								title="Delete stock"
							>
								<DeleteIcon />
							</button>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{:else}
		<p>No stocks found.</p>
	{/if}
	<table>
		<tbody>
			<tr>
				<td><input type="text" placeholder="WKN" bind:value={newStockWkn} /></td>
				<td><input type="text" placeholder="Ticker" bind:value={newStockTicker} /></td>
				<td>
					<input type="number" step="0.01" placeholder="Price" bind:value={newStockPriceEuros} />
				</td>
				<td></td>
				<td><button onclick={createStock}>Add Stock</button></td>
			</tr>
		</tbody>
	</table>
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
	.icon-button {
		width: auto;
		background: none;
		border: none;
		padding: 0.25rem;
		margin: 0 0.25rem;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		vertical-align: middle;
	}
	.save-button {
		color: var(--pico-color-green-500);
	}
	.refresh-button {
		color: var(--pico-color-blue-500);
	}
	.refresh-button:disabled {
		color: var(--pico-muted-color);
		cursor: not-allowed;
	}
	.delete-button {
		color: var(--pico-del-color);
	}
	.delete-button:disabled {
		color: var(--pico-muted-color);
		cursor: not-allowed;
	}
	:global(html[data-theme='dark']) .save-button {
		color: var(--pico-color-green-350);
	}
	:global(html[data-theme='dark']) .refresh-button {
		color: var(--pico-color-blue-350);
	}
</style>
