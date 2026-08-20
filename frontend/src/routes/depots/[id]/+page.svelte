<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { formatCurrency } from '$lib/utils';
	import type { Lot, Position, TradeDTO, TradeType } from '$lib/types';

	let { data } = $props();

	let tradeType = $state<TradeType>('BUY');
	let wkn = $state('');
	let quantity = $state<number | null>(null);
	let total = $state<number | null>(null);
	let date = $state(new Date().toISOString().split('T')[0]);
	let isSubmitting = $state(false);
	let errorMessage = $state('');

	const hasPositions = $derived(data.portfolio.positions.length > 0);
	const hasTrades = $derived(data.trades.length > 0);

	const heldWKNs = $derived(data.portfolio.positions.map((p: Position) => p.wkn));
	let expandedWKNs = $state<string[]>([]);

	function isExpanded(wknToCheck: string) {
		return expandedWKNs.includes(wknToCheck);
	}

	function toggleLots(wknToToggle: string) {
		expandedWKNs = isExpanded(wknToToggle)
			? expandedWKNs.filter((w) => w !== wknToToggle)
			: [...expandedWKNs, wknToToggle];
	}

	function pricePerShare(totalInCents: number, shares: number) {
		return shares > 0 ? formatCurrency(Math.round(totalInCents / shares)) : '-';
	}

	function lotLabel(lot: Lot) {
		return lot.remaining === lot.quantity
			? `${lot.quantity}`
			: `${lot.remaining} of ${lot.quantity}`;
	}

	async function submitTrade(event: Event) {
		event.preventDefault();
		errorMessage = '';

		if (!quantity || !total) {
			errorMessage = 'Quantity and total are required.';
			return;
		}

		isSubmitting = true;
		const res = await fetch(`/api/depots/${data.depot.id}/trades`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				wkn,
				type: tradeType,
				quantity,
				totalInCents: Math.round(total * 100),
				timestamp: new Date(date).toISOString()
			})
		});
		isSubmitting = false;

		if (!res.ok) {
			errorMessage = (await res.text()) || 'Could not save the trade.';
			return;
		}

		wkn = '';
		quantity = null;
		total = null;
		await invalidateAll();
	}

	async function deleteTrade(id: number) {
		if (!confirm('Delete this trade? Its wallet transaction is removed as well.')) return;

		errorMessage = '';
		const res = await fetch(`/api/trades/${id}`, { method: 'DELETE' });
		if (!res.ok) {
			errorMessage = (await res.text()) || 'Could not delete the trade.';
			return;
		}
		await invalidateAll();
	}
</script>

<article>
	<header>
		<strong>{data.depot.name}</strong>
		{#if data.wallet}
			<span class="muted"> settled via {data.wallet.name}</span>
		{/if}
		{#if data.budget}
			<span class="muted"> · budget {data.budget.name}</span>
		{/if}
	</header>
	<div class="totals">
		<div>
			<small>Invested</small>
			<strong>{formatCurrency(data.portfolio.investedInCents)}</strong>
		</div>
		<div>
			<small>Realized</small>
			<strong
				class:gain={data.portfolio.realizedGainInCents > 0}
				class:loss={data.portfolio.realizedGainInCents < 0}
			>
				{formatCurrency(data.portfolio.realizedGainInCents)}
			</strong>
		</div>
	</div>
</article>

<article>
	<header><strong>{tradeType === 'BUY' ? 'Buy' : 'Sell'}</strong></header>
	<form onsubmit={submitTrade}>
		<div class="grid">
			<label>
				Type
				<select bind:value={tradeType}>
					<option value="BUY">Buy</option>
					<option value="SELL">Sell</option>
				</select>
			</label>
			<label>
				WKN
				<input type="text" list="wkn-options" bind:value={wkn} autocomplete="on" required />
				<datalist id="wkn-options">
					{#each heldWKNs as option (option)}
						<option value={option}></option>
					{/each}
				</datalist>
			</label>
		</div>
		<div class="grid">
			<label>
				Quantity
				<input type="number" step="0.000001" bind:value={quantity} required />
			</label>
			<label>
				Total (€)
				<input type="number" step="0.01" bind:value={total} required />
			</label>
			<label>
				Date
				<input type="date" bind:value={date} required />
			</label>
		</div>
		{#if errorMessage}
			<p style="color: var(--pico-del-color);">{errorMessage}</p>
		{/if}
		<button type="submit" aria-busy={isSubmitting}>
			{tradeType === 'BUY' ? 'Buy shares' : 'Sell shares'}
		</button>
	</form>
</article>

<article>
	<header><strong>Positions</strong></header>
	{#if hasPositions}
		<table>
			<thead>
				<tr>
					<th>Position</th>
					<th>Quantity</th>
					<th>Average price</th>
					<th>Invested</th>
					<th>Current Value</th>
					<th>+/-</th>
				</tr>
			</thead>
			<tbody>
				{#each data.portfolio.positions as position (position.wkn)}
					<tr>
						<td>
							<button
								class="toggle"
								aria-expanded={isExpanded(position.wkn)}
								onclick={() => toggleLots(position.wkn)}
							>
								<span class="caret">{isExpanded(position.wkn) ? '▾' : '▸'}</span>
								{position.wkn}
							</button>
						</td>
						<td>{position.quantity}</td>
						<td>{formatCurrency(position.avgPriceInCents)}</td>
						<td>{formatCurrency(position.investedInCents)}</td>
						<td>{formatCurrency(position.currentValueInCents ?? 0)}</td>
						<td
							class:gain={(position.unrealizedGainInCents ?? 0) > 0}
							class:loss={(position.unrealizedGainInCents ?? 0) < 0}
						>
							{formatCurrency(position.unrealizedGainInCents ?? 0)}
						</td>
					</tr>
					{#if isExpanded(position.wkn)}
						{#each position.lots as lot (lot.tradeId)}
							<tr class="lot">
								<td>{new Date(lot.dateOfPurchase).toLocaleDateString('de-DE')}</td>
								<td>{lotLabel(lot)}</td>
								<td>{pricePerShare(lot.totalInCents, lot.quantity)}</td>
								<td>{formatCurrency(lot.remainingCostInCents)}</td>
								<td></td>
								<td></td>
							</tr>
						{/each}
					{/if}
				{/each}
			</tbody>
		</table>
		<small class="muted">
			Click a position to show its lots, oldest first - that is the order a sale consumes them.
		</small>
	{:else}
		<p>No shares held in this depot.</p>
	{/if}
</article>

<article>
	<header><strong>Trades</strong></header>
	{#if hasTrades}
		<table>
			<thead>
				<tr>
					<th>Date</th>
					<th>Type</th>
					<th>WKN</th>
					<th>Quantity</th>
					<th>Total</th>
					<th>Realized</th>
					<th></th>
				</tr>
			</thead>
			<tbody>
				{#each data.trades as trade (trade.id)}
					<tr>
						<td>{new Date(trade.timestamp).toLocaleDateString('de-DE')}</td>
						<td>{trade.type}</td>
						<td>{trade.wkn}</td>
						<td>{trade.quantity}</td>
						<td>{formatCurrency(trade.totalInCents)}</td>
						<td
							class:gain={trade.realizedGainInCents > 0}
							class:loss={trade.realizedGainInCents < 0}
						>
							{trade.type === 'SELL' ? formatCurrency(trade.realizedGainInCents) : '-'}
						</td>
						<td title={trade.canDelete ? '' : 'Shares from this buy have already been sold'}>
							<button
								class="secondary outline"
								disabled={!trade.canDelete}
								onclick={() => deleteTrade(trade.id)}
							>
								Delete
							</button>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{:else}
		<p>No trades yet.</p>
	{/if}
</article>

<style>
	.muted {
		color: var(--pico-muted-color);
		font-weight: normal;
	}

	.totals {
		display: flex;
		gap: 2rem;
	}

	.totals div {
		display: flex;
		flex-direction: column;
	}

	.toggle {
		background: none;
		border: none;
		padding: 0;
		margin: 0;
		width: auto;
		color: inherit;
		font-weight: inherit;
		text-align: left;
	}

	.caret {
		color: var(--pico-muted-color);
		margin-right: 0.25rem;
	}

	.lot td {
		color: var(--pico-muted-color);
		padding-left: 1.5rem;
		font-size: 0.9em;
	}

	.gain {
		color: var(--pico-color-green-500);
	}

	.loss {
		color: var(--pico-del-color);
	}

	button {
		padding: 0.25rem 0.75rem;
		margin-bottom: 0;
	}
</style>
