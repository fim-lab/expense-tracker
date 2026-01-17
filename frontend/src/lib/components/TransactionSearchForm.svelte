<script lang="ts">
	let { criteria, budgets, wallets } = $props();
</script>

<form>
	<div class="grid">
		<label for="search">
			Description
			<input type="search" id="search" name="search" bind:value={criteria.q} />
		</label>
		<div></div>
	</div>

	<div class="grid">
		<label for="from">
			From
			<input type="date" id="from" name="from" bind:value={criteria.from} />
		</label>
		<label for="until">
			Until
			<input type="date" id="until" name="until" bind:value={criteria.until} />
		</label>
	</div>
	<div class="grid">
		<label for="budget">
			Budget
			<select id="budget" name="budget" bind:value={criteria.budget_id}>
				<option value={undefined}>All</option>
				{#each budgets as budget}
					<option value={budget.id}>{budget.name}</option>
				{/each}
			</select>
		</label>
		<label for="wallet">
			Wallet
			<select id="wallet" name="wallet" bind:value={criteria.wallet_id}>
				<option value={undefined}>All</option>
				{#each wallets as wallet}
					<option value={wallet.id}>{wallet.name}</option>
				{/each}
			</select>
		</label>
		<label for="type">
			Type
			<select id="type" name="type" bind:value={criteria.type}>
				<option value={undefined}>All</option>
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
