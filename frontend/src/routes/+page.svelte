<script lang="ts">
	import type { PageData } from './$types';
	import TransactionTable from '$lib/components/TransactionTable.svelte';

	let { data } = $props();

	// We initialize a reactive state with the data from the server load
	let transactions = $state(data.transactions);
</script>

<div class="grid">
	<aside>
		<article>
			<header><strong>Wallets</strong></header>
			<ul>
				{#each data.wallets as wallet}
					<li>{wallet.name}</li>
				{/each}
			</ul>
		</article>
	</aside>

	<section>
		<TransactionTable bind:transactions />
	</section>
</div>
