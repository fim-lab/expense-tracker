<script lang="ts">
	import { page } from '$app/state';
	import type { TransactionType } from '$lib/types';
	import { updateParams, debounce } from '$lib/utils';

	let { budgets = [], wallets = [] } = $props();
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
	const debtOnly = $derived(page.url.searchParams.get('debt') === 'true');

	const debouncedUpdateParams = debounce((value: string) => {
		updateParams({ q: value });
	}, 500);

	const currentMonthLabel = new Date().toLocaleDateString('de-DE', { month: 'long' });
	const currentYearLabel = String(new Date().getFullYear());

	function formatDateISO(date: Date): string {
		const year = date.getFullYear();
		const month = String(date.getMonth() + 1).padStart(2, '0');
		const day = String(date.getDate()).padStart(2, '0');
		return `${year}-${month}-${day}`;
	}

	function setCurrentMonth() {
		const now = new Date();
		updateParams({ from: formatDateISO(new Date(now.getFullYear(), now.getMonth(), 1)), until: undefined });
	}

	function setMonthToDate() {
		const now = new Date();
		updateParams({ from: formatDateISO(new Date(now.getFullYear(), now.getMonth() - 1, now.getDate())), until: undefined });
	}

	function setCurrentYear() {
		const now = new Date();
		updateParams({ from: formatDateISO(new Date(now.getFullYear(), 0, 1)), until: undefined });
	}

	function setYearToDate() {
		const now = new Date();
		updateParams({ from: formatDateISO(new Date(now.getFullYear() - 1, now.getMonth(), now.getDate())), until: undefined });
	}
</script>

<form>
	<div class="grid">
		<label for="search">
			Description
			<input type="search" value={searchTerm} id="search" name="search" 
				oninput={(e) => debouncedUpdateParams(e.currentTarget.value)} />
		</label>
		<div></div>
	</div>

	<div class="grid quick-range-buttons">
		<button type="button" class="secondary outline" onclick={setCurrentMonth}>{currentMonthLabel}</button>
		<button type="button" class="secondary outline" onclick={setMonthToDate}>MTD</button>
		<button type="button" class="secondary outline" onclick={setCurrentYear}>{currentYearLabel}</button>
		<button type="button" class="secondary outline" onclick={setYearToDate}>YTD</button>
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
	<label for="debt">
		<input type="checkbox" id="debt" name="debt" checked={debtOnly}
			onchange={(e) => updateParams({ debt: e.currentTarget.checked ? 'true' : undefined })} />
		Debt only
	</label>
</form>

<style>
	form {
		margin-bottom: 2rem;
	}

	.quick-range-buttons button {
		padding: 0.25rem 0.75rem;
		font-size: 0.85rem;
	}
</style>
