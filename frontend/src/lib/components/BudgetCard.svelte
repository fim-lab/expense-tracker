<script lang="ts">
	import { formatCurrency } from '$lib/utils';

	let { budget } = $props();

	const percentage = $derived(() => {
		if (budget.limitCents <= 0) return 0;
		const p = (budget.balanceCents / budget.limitCents) * 100;
		return Math.max(0, Math.min(100, p));
	});

	const isOverBudget = $derived(percentage() > 100);
</script>

<div class="card">
	<div class="card-info">
		<p class="card-title">
			{budget.name}
			<span class="card-amount"
				>{formatCurrency(budget.balanceCents)} / {formatCurrency(budget.limitCents)}</span
			>
		</p>
		<progress value={percentage()} max="100" class:over-budget={isOverBudget}></progress>
	</div>
</div>

<style>
	.card {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem;
		margin-bottom: 0.75rem;
		background: var(--pico-card-background-color);
		border-radius: var(--pico-border-radius);
		box-shadow: var(--pico-card-box-shadow);
		border-left: 4px solid var(--pico-primary);
	}

	.card-info {
		width: 100%;
	}

	.card-title {
		font-weight: bold;
		margin-bottom: 0.5rem;
		display: flex;
		justify-content: space-between;
	}

	.card-amount {
		font-weight: normal;
		font-size: 0.9em;
		font-family: ui-monospace, SFMono-Regular, monospace;
	}

	progress {
		width: 100%;
		height: 0.75rem;
		border-radius: var(--pico-border-radius);
	}

	progress.over-budget::-webkit-progress-value {
		background-color: var(--pico-del-color);
	}
	progress.over-budget::-moz-progress-bar {
		background-color: var(--pico-del-color);
	}
</style>
