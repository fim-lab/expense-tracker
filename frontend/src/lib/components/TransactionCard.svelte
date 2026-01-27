<script lang="ts">
	import { formatCurrency } from '$lib/utils';

	let { transaction, ondelete } = $props();
	const isExpense = $derived(transaction.type === 'EXPENSE');
</script>

<div class="tx-card" class:expense={isExpense}>
	<div class="tx-info">
		<p class="tx-title">
			{transaction.description}
			<span class="tx-amount"
				>({isExpense ? '-' : '+'}{formatCurrency(transaction.amountInCents)})</span
			>
		</p>
		<p class="tx-meta">
			{transaction.budgetName}
			{transaction.budgetName ? ' • ' : ''}
			{transaction.walletName}
		</p>
		<p class="tx-date">{new Date(transaction.date).toLocaleDateString('de-DE')}</p>
	</div>

	<button class="delete-icon" onclick={() => ondelete(transaction.id)} aria-label="Delete">
		<svg
			xmlns="http://www.w3.org/2000/svg"
			width="20"
			height="20"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			stroke-linecap="round"
			stroke-linejoin="round"
		>
			<path d="M3 6h18"></path>
			<path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path>
			<path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path>
		</svg>
	</button>
</div>

<style>
	.tx-card {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem;
		margin-bottom: 0.75rem;
		background: var(--pico-card-background-color);
		border-radius: var(--pico-border-radius);
		box-shadow: var(--pico-card-box-shadow);
		border-left: 4px solid var(--pico-ins-color);
	}

	.tx-card.expense {
		border-left-color: var(--pico-del-color);
	}

	.tx-title {
		font-weight: bold;
		margin-bottom: 0;
	}

	.tx-amount {
		font-weight: normal;
		font-size: 0.9em;
	}

	.tx-meta {
		font-size: 0.75rem;
		color: var(--pico-muted-color);
		text-transform: uppercase;
		letter-spacing: 0.05rem;
		margin-bottom: 0;
	}

	.tx-date {
		font-size: 0.75rem;
		color: var(--pico-muted-color);
		margin-bottom: 0;
	}

	.delete-icon {
		background: transparent;
		border: none;
		color: var(--pico-del-color);
		padding: 0.5rem;
		margin: 0;
		width: auto;
		opacity: 0.6;
		transition: opacity 0.2s;
	}

	.delete-icon:hover {
		background: transparent;
		opacity: 1;
		color: var(--pico-del-color);
	}
</style>
