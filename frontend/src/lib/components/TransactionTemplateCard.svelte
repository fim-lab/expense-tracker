<script lang="ts">
	import SaveIcon from '$lib/components/icons/SaveIcon.svelte';
	import DeleteIcon from '$lib/components/icons/DeleteIcon.svelte';

	let {
		template,
		ondelete,
		onuse,
		onsave = undefined,
		onamountchange = undefined,
		draggable = false,
		dragging = false,
		ondragstart = undefined,
		ondragend = undefined,
		ondragover = undefined,
		ondrop = undefined
	} = $props();
	const isExpense = $derived(template.type === 'EXPENSE');
	let amountEuros = $state(template.amountInCents / 100);

	function handleAmountInput() {
		onamountchange?.(template.id, Math.round(amountEuros * 100));
	}
</script>

<div
	class="tx-card"
	class:expense={isExpense}
	class:dragging
	{draggable}
	ondragstart={(e) => ondragstart?.(e, template)}
	ondragend={(e) => ondragend?.(e, template)}
	ondragover={(e) => ondragover?.(e, template)}
	ondrop={(e) => ondrop?.(e, template)}
>
	<div class="tx-info">
		<p class="tx-title">
			{template.description}
			<span class="tx-amount-edit">
				({isExpense ? '-' : '+'}
				<input
					class="amount-input"
					type="number"
					step="0.01"
					bind:value={amountEuros}
					oninput={handleAmountInput}
				/>)
			</span>
		</p>
		<p class="tx-meta">
			{template.budgetName}
			{template.budgetName ? ' • ' : ''}
			{template.walletName}
		</p>
	</div>

	<div class="tx-actions">
		<button
			class="action-icon save-icon"
			onclick={() => onsave?.(template.id, Math.round(amountEuros * 100))}
			aria-label="Save as default"
			title="Save as default amount"
		>
			<SaveIcon />
		</button>
		<button
			class="action-icon use-icon"
			onclick={() => onuse(template)}
			aria-label="Use"
			title="Create transaction now"
		>
			<svg
				xmlns="http://www.w3.org/2000/svg"
				width="24"
				height="24"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
			>
				<path d="M12 5v14" />
				<path d="M5 12h14" />
			</svg>
		</button>
		<button
			class="action-icon delete-icon"
			onclick={() => ondelete(template.id)}
			aria-label="Delete"
			title="Delete template"
		>
			<DeleteIcon />
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

	.tx-card.dragging {
		opacity: 0.4;
	}

	.tx-title {
		font-weight: bold;
		margin-bottom: 0;
	}

	.tx-amount-edit {
		font-weight: normal;
		font-size: 0.9em;
	}

	.amount-input {
		display: inline-block;
		width: 5.5rem;
		margin: 0 0.15rem;
		padding: 0.15rem 0.3rem;
	}

	.tx-meta {
		font-size: 0.75rem;
		color: var(--pico-muted-color);
		text-transform: uppercase;
		letter-spacing: 0.05rem;
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
	.use-icon {
		color: var(--pico-ins-color);
	}
	.use-icon:hover {
		color: var(--pico-ins-color);
	}
	.save-icon {
		color: var(--pico-color-green-500);
	}
	.save-icon:hover {
		color: var(--pico-color-green-500);
	}
	:global(html[data-theme='dark']) .save-icon {
		color: var(--pico-color-green-350);
	}
</style>
