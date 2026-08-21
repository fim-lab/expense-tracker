<script lang="ts">
	import { formatCurrency } from '$lib/utils';

	let { amountInCents } = $props();
</script>

<div class="card">
	<div class="card-info">
		<p class="card-title">
			Debt
			<span class="card-amount">{formatCurrency(amountInCents)}</span>
		</p>
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
		border-left: 4px solid var(--pico-del-color);
	}

	.card-info {
		width: 100%;
	}

	.card-title {
		font-weight: bold;
		margin-bottom: 0;
		display: flex;
		justify-content: space-between;
	}

	.card-amount {
		font-weight: normal;
		font-size: 1em;
		font-family: ui-monospace, SFMono-Regular, monospace;
	}
</style>
