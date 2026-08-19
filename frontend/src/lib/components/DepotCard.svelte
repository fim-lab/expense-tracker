<script lang="ts">
	import { formatCurrency } from '$lib/utils';

	let { depot, href = '', subtitle = '' } = $props();
</script>

<div class="card">
	<div class="card-info">
		<p class="card-title">
			{#if href}
				<a {href}>{depot.name}</a>
			{:else}
				{depot.name}
			{/if}
			<span class="card-amount">{formatCurrency(depot.investedInCents ?? 0)}</span>
		</p>
		{#if subtitle}
			<small class="card-subtitle">{subtitle}</small>
		{/if}
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
		border-left: 4px solid var(--pico-primary-inverse);
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

	.card-subtitle {
		color: var(--pico-muted-color);
	}
</style>
