<script lang="ts">
	import { page } from '$app/state';
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

	<div class="tx-actions">
		<a
			href={`/transactions/update/${transaction.id}?redirect=${encodeURIComponent(page.url.search)}`}
			class="action-icon update-icon"
			aria-label="Update"
		>
			<svg
				xmlns="http://www.w3.org/2000/svg"
				width="20"
				height="20"
				viewBox="0 0 24 24"
				fill="none"
				stroke="black"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
			>
				<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
				<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
			</svg>
		</a>
		<button class="action-icon delete-icon" onclick={() => ondelete(transaction.id)} aria-label="Delete">
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

	.tx-actions {
		display: flex;
		gap: 0.5rem;
	}

	.action-icon {
		background: transparent;
		border: none;
		padding: 0.5rem;
		margin: 0;
		width: auto;
		opacity: 0.6;
		transition: opacity 0.2s;
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}

	.action-icon:hover {
		background: transparent;
		opacity: 1;
	}

	.delete-icon {
		color: var(--pico-del-color);
	}
	.delete-icon:hover {
		color: var(--pico-del-color);
	}
	.update-icon {
		color: var(--pico-primary-inverse);
	}
	.update-icon:hover {
		color: var(--pico-primary-inverse);
	}
</style>
