<script lang="ts">
	import type { PageData, ActionData } from './$types';

	export let data: PageData;
	export let form: ActionData;

	const { transaction, wallets, budgets } = data;

	const formattedDate = new Date(transaction.date).toISOString().split('T')[0];
</script>

<article>
	<h3>Update Transaction</h3>
	<form method="POST">
		<input type="hidden" name="id" value={transaction.id} />

		<div class="grid">
			<label>
				Date
				<input type="date" name="date" value={formattedDate} required />
			</label>
			<label>
				Type
				<select name="type" value={transaction.type}>
					<option value="EXPENSE">Expense</option>
					<option value="INCOME">Income</option>
				</select>
			</label>
		</div>

		<label>
			Description
			<input
				type="text"
				name="description"
				value={transaction.description}
				placeholder="Grocery shopping..."
				required
			/>
		</label>

		<div class="grid">
			<label>
				Amount (EUR)
				<input
					type="number"
					step="0.01"
					name="amount"
					value={transaction.amountInCents / 100}
					required
				/>
			</label>
			<label>
				Wallet
				<select name="walletId" required>
					{#each wallets as wallet}
						<option value={wallet.id} selected={wallet.id === transaction.walletId}>
							{wallet.name}
						</option>
					{/each}
				</select>
			</label>
		</div>

		<label>
			Budget
			<select name="budgetId" required>
				{#each budgets as budget}
					<option value={budget.id} selected={budget.id === transaction.budgetId}>
						{budget.name}
					</option>
				{/each}
			</select>
		</label>

		{#if form?.error}
			<p class="error-message">{form.error}</p>
		{/if}

		<button type="submit">Update Transaction</button>
	</form>
</article>

<style>
	.error-message {
		color: var(--pico-del-color);
		margin-top: 1rem;
		margin-bottom: 0;
	}
</style>
