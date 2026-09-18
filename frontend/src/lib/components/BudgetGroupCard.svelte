<script lang="ts">
	import BudgetCard from '$lib/components/BudgetCard.svelte';
	import { computeBudgetProgress } from '$lib/budgetProgress';
	import { formatCurrency } from '$lib/utils';
	import type { Budget, BudgetGroup } from '$lib/types';

	let { group, budgets }: { group: BudgetGroup; budgets: Budget[] } = $props();

	let expanded = $state(false);

	const totalBalanceCents = $derived(budgets.reduce((sum, b) => sum + b.balanceCents, 0));
	const totalLimitCents = $derived(budgets.reduce((sum, b) => sum + b.limitCents, 0));
	const progress = $derived(computeBudgetProgress(totalBalanceCents, totalLimitCents));
</script>

<button
	type="button"
	class="card group-card"
	aria-expanded={expanded}
	onclick={() => (expanded = !expanded)}
>
	<div class="card-info">
		<p class="card-title">
			<span><span class="caret">{expanded ? '▾' : '▸'}</span> {group.name}</span>
			<span class="card-amount"
				>{formatCurrency(totalBalanceCents)} / {formatCurrency(totalLimitCents)}</span
			>
		</p>
		<div
			class="progress-track"
			role="progressbar"
			aria-valuenow={Math.round(progress.displayPercentage)}
			aria-valuemin="0"
			aria-valuemax="100"
		>
			{#if progress.isOverBudget}
				<div class="progress-fill within-limit" style="width: {progress.withinLimitWidth}%"></div>
				<div class="progress-fill overflow" style="width: {progress.overflowWidth}%"></div>
			{:else}
				<div class="progress-fill lower-limit" style="width: {progress.lowerLimitWidth}%"></div>
				{#if progress.isOverLowerLimit}
					<div
						class="progress-fill"
						style="width: {progress.displayPercentage - progress.lowerLimitWidth}%"
					></div>
				{/if}
			{/if}
		</div>
	</div>
</button>
{#if expanded}
	<div class="group-members">
		{#each budgets as budget (budget.id)}
			<BudgetCard {budget} />
		{/each}
	</div>
{/if}

<style>
	.card {
		display: flex;
		justify-content: space-between;
		align-items: center;
		width: 100%;
		padding: 1rem;
		margin-bottom: 0.75rem;
		background: var(--pico-card-background-color);
		border-radius: var(--pico-border-radius);
		box-shadow: var(--pico-card-box-shadow);
		border-left: 4px solid var(--pico-primary);
		color: inherit;
		text-align: left;
		cursor: pointer;

		p {
			color: inherit;
		}
	}

	.group-card {
		border: none;
		border-left: 4px solid var(--pico-primary);
		font: inherit;
	}

	.caret {
		color: var(--pico-muted-color);
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

	.progress-track {
		display: flex;
		width: 100%;
		height: 0.75rem;
		border-radius: var(--pico-border-radius);
		background-color: var(--pico-progress-background-color);
		overflow: hidden;
	}

	.progress-fill {
		height: 100%;
		background-color: var(--pico-progress-color);
	}

	.progress-fill.within-limit {
		border-right: 2px solid #000;
	}

	.progress-fill.overflow {
		background-color: var(--pico-color-green-500);
	}

	.progress-fill.lower-limit {
		background-color: var(--pico-color-red-600);
	}

	:global(html[data-theme='dark']) .progress-fill.overflow {
		background-color: var(--pico-color-green-350);
	}
	:global(html[data-theme='dark']) .progress-fill.lower-limit {
		background-color: var(--pico-color-red-350);
	}

	.group-members {
		margin-left: 1rem;
		padding-left: 0.75rem;
		border-left: 2px solid var(--pico-muted-border-color);
	}
</style>
