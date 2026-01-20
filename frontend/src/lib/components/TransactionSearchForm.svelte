<script lang="ts">
	import { page } from '$app/state';
	import type { TransactionType } from '$lib/types';
	import { updateParams } from '$lib/utils';

	let { budgets, wallets } = $props();
	const searchTerm = $derived(
		page.url.searchParams.get('q') ?? ''
	);
	const from = $derived(
		page.url.searchParams.get('from') ?? ''
	);
	const until = $derived(
		page.url.searchParams.get('until') ?? ''
	);
	const budgetId = $derived(() => {
		const id = page.url.searchParams.get('budgetId');
		return id ? Number(id) : 'all';
	});
	const walletId = $derived(() => {
		const id = page.url.searchParams.get('walletId');
		return id ? Number(id) : 'all';
	});
	const type = $derived(() => {
		const typeParam = page.url.searchParams.get('type');
		return (typeParam === 'INCOME' || typeParam === 'EXPENSE') ? typeParam : 'all';
	});
</script>

<form>
	<div class="grid">
		<label for="search">
			Description
			<input type="search" value={searchTerm} id="search" name="search" 
				oninput={(e) => updateParams({ q: e.currentTarget.value})} />
		</label>
		<div></div>
	</div>

	<div class="grid">
		<label for="from">
			From
			<input type="date" value={from} id="from" name="from" oninput={(e) => updateParams({ from: e.currentTarget.value})} />
		</label>
		<label for="until">
			Until
			<input type="date" value={until} id="until" name="until" oninput={(e) => updateParams({ until: e.currentTarget.value})} />
		</label>
	</div>
	<div class="grid">
		<label for="budget">
			Budget
			<select id="budget" name="budget" onchange={(e) => updateParams({ budget_id: e.currentTarget.value === 'all' ? undefined : Number(e.currentTarget.value)})} value={budgetId}>
				<option value={'all'}>All</option>
				{#each budgets as budget}
					<option value={budget.id}>{budget.name}</option>
				{/each}
			</select>
		</label>
		<label for="wallet">
			Wallet
			<select id="wallet" name="wallet" onchange={(e) => updateParams({ wallet_id: e.currentTarget.value === 'all' ? undefined : Number(e.currentTarget.value)})} value={walletId}>
				<option value={'all'}>All</option>
				{#each wallets as wallet}
					<option value={wallet.id}>{wallet.name}</option>
				{/each}
			</select>
		</label>
		<label for="type">
			Type
			<select id="type" name="type" onchange={(e) => updateParams({ type: e.currentTarget.value === 'all' ? undefined : e.currentTarget.value as TransactionType})} value={type}>
				<option value={'all'}>All</option>
				<option value="INCOME">Income</option>
				<option value="EXPENSE">Expense</option>
			</select>
		</label>
	</div>
</form>

<style>
	form {
		margin-bottom: 2rem;
	}
</style>
