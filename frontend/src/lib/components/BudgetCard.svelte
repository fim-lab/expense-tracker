<script lang="ts">
	import { page } from '$app/state';
	import { formatCurrency } from '$lib/utils';

	let { budget } = $props();

	const rawPercentage = $derived(
		budget.limitCents <= 0 ? 0 : (budget.balanceCents / budget.limitCents) * 100
	);

	const isOverBudget = $derived(rawPercentage > 100);
	const isOverLowerLimit = $derived(rawPercentage > 10);

	const displayPercentage = $derived(Math.max(0, Math.min(100, rawPercentage)));

	const withinLimitWidth = $derived(isOverBudget ? 10000 / rawPercentage : displayPercentage);
	const overflowWidth = $derived(isOverBudget ? 100 - withinLimitWidth : 0);
	const lowerLimitWidth = $derived(Math.min(10, displayPercentage));

	const href = $derived.by(() => {
		const params = new URLSearchParams(page.url.searchParams);
		params.set('budget_id', String(budget.id));
		params.set('page', '1');
		return `/?${params}`;
	});
</script>

<a {href} class="card" data-sveltekit-noscroll>
	<div class="card-info">
		<p class="card-title">
			{budget.name}
			<span class="card-amount"
				>{formatCurrency(budget.balanceCents)} / {formatCurrency(budget.limitCents)}</span
			>
		</p>
		<div
			class="progress-track"
			role="progressbar"
			aria-valuenow={Math.round(displayPercentage)}
			aria-valuemin="0"
			aria-valuemax="100"
		>
			{#if isOverBudget}
				<div class="progress-fill within-limit" style="width: {withinLimitWidth}%"></div>
				<div class="progress-fill overflow" style="width: {overflowWidth}%"></div>
			{:else}
				<div class="progress-fill lower-limit" style="width: {lowerLimitWidth}%"></div>
				{#if isOverLowerLimit}
					<div class="progress-fill" style="width: {displayPercentage - lowerLimitWidth}%"></div>
				{/if}
			{/if}
		</div>
	</div>
</a>

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
		color: inherit;
		text-decoration: none;

		p {
			color: inherit;
		}
	}

	.card:hover,
	.card:focus,
	.card:active {
		color: inherit;
		text-decoration: none;
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

	@media (prefers-color-scheme: dark) {
		.progress-fill.overflow {
			background-color: var(--pico-color-green-350);
		}
		.progress-fill.lower-limit {
			background-color: var(--pico-color-red-350);
		}
	}
</style>
