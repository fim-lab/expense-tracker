<script lang="ts">
	import { formatCurrency, getShortMonthYearString } from '$lib/utils';
	import type { Budget, User, Wallet } from '$lib/types';
	import { goto } from '$app/navigation';

	let { data } = $props();

	let user = $state<User | null>(
		data.user
			? {
					...data.user,
					isEditing: false,
					newSalaryEuros: data.user.salaryCents / 100
				}
			: null
	);

	let budgets = $state<Budget[]>(
		data.budgets.map((b: Budget) => ({
			...b,
			isEditing: false,
			newLimitEuros: b.limitCents / 100
		}))
	);

	let wallets: Wallet[] = data.wallets;
	let selectedWalletId = $state<number | undefined>(wallets.length > 0 ? wallets[0].id : undefined);

	let transactionDescription = $state<string>('');

	$effect(() => {
		const dateForTx = new Date();
		const dayOfMonth = dateForTx.getDate();

		if (dayOfMonth <= 10) {
			dateForTx.setMonth(dateForTx.getMonth() - 1);
		}

		transactionDescription = `Gehalt ${getShortMonthYearString(dateForTx)}`;
	});

	$effect(() => {
		if (budgets) {
			calculateTotalBudgetLimit();
		}
		if (user) {
			updateSalaryInput();
		}
	});

	let totalBudgetLimitCents = $state(0);
	let currentSalaryInput = $state(0);
	let salaryMismatch = $state(0);

	function calculateTotalBudgetLimit() {
		let total = 0;
		for (const budget of budgets) {
			total +=
				budget.newLimitEuros !== undefined
					? Math.round(budget.newLimitEuros * 100)
					: budget.limitCents;
		}
		totalBudgetLimitCents = total;
		calculateSalaryMismatch();
	}

	function updateSalaryInput() {
		if (user && user.salaryCents !== undefined) {
			currentSalaryInput =
				user.newSalaryEuros !== undefined
					? Math.round(user.newSalaryEuros * 100)
					: user.salaryCents;
		} else {
			salaryMismatch = 0;
		}
		calculateSalaryMismatch();
	}

	function calculateSalaryMismatch() {
		salaryMismatch = totalBudgetLimitCents - currentSalaryInput;
	}

	function fixSalary() {
		if (user) {
			user.newSalaryEuros = totalBudgetLimitCents / 100;
			updateSalaryInput();
		}
	}

	function fixBudget(budget: Budget) {
		if (budget.newLimitEuros !== undefined) {
			budget.newLimitEuros = (Math.round(budget.newLimitEuros * 100) - salaryMismatch) / 100;
			calculateTotalBudgetLimit();
		}
	}

	async function generateSalaryTransactions() {
		if (salaryMismatch !== 0) {
			alert(
				'The sum of your budget limits does not equal your salary. Please fix the mismatch before generating transactions.'
			);
			return;
		}

		if (selectedWalletId === undefined) {
			alert('Please select a wallet for the transactions.');
			return;
		}

		if (!transactionDescription.trim()) {
			alert('Please enter a description for the transactions.');
			return;
		}

		const today = new Date().toISOString();
		const requests = [];

		for (const budget of budgets) {
			if (budget.newLimitEuros !== undefined && budget.newLimitEuros > 0) {
				const payload = {
					date: today,
					description: transactionDescription,
					amountInCents: Math.round(budget.newLimitEuros * 100),
					walletId: selectedWalletId,
					budgetId: budget.id,
					type: 'INCOME',
					isPending: false,
					tags: []
				};
				requests.push(
					fetch(`/api/transactions`, {
						method: 'POST',
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify(payload)
					})
				);
			}
		}

		if (requests.length === 0) {
			alert('No transactions to generate. Please ensure budgets have a positive limit.');
			return;
		}

		try {
			const responses = await Promise.all(requests);
			const allOk = responses.every((res) => res.ok);

			if (allOk) {
				alert('Salary transactions generated successfully!');
				goto('/');
			} else {
				alert('Some transactions failed to generate. Please check the console for details.');
				console.error('Failed responses:', responses);
			}
		} catch (error) {
			console.error('Error sending transactions:', error);
			alert('An error occurred while generating transactions.');
		}
	}
</script>

<h1>Generate Salary Transactions</h1>

{#if salaryMismatch !== 0}
	<article class="warning">
		The sum of your budget limits ({formatCurrency(totalBudgetLimitCents)}) does not equal your
		salary ({formatCurrency(currentSalaryInput)}).<br />
		The Difference is {formatCurrency(salaryMismatch)}
	</article>
{/if}

{#if user && budgets.length > 0 && wallets.length > 0}
	<div>
		<label for="wallet-select">Select Wallet:</label>
		<select id="wallet-select" bind:value={selectedWalletId}>
			{#each wallets as wallet (wallet.id)}
				<option value={wallet.id}>{wallet.name}</option>
			{/each}
		</select>
	</div>

	<div>
		<label for="transaction-description">Transaction Description:</label>
		<input type="text" id="transaction-description" bind:value={transactionDescription} />
	</div>

	<table>
		<tbody>
			<tr>
				<td>Salary</td>
				<td>
					<input
						type="number"
						step="0.01"
						bind:value={user.newSalaryEuros}
						oninput={updateSalaryInput}
					/>
				</td>
				<td>
					{#if salaryMismatch !== 0}
						<button class="secondary" onclick={fixSalary}>Fix</button>
					{/if}
				</td>
			</tr>
			{#each budgets as budget (budget.id)}
				<tr>
					<td>{budget.name}</td>
					<td>
						<input
							type="number"
							step="0.01"
							bind:value={budget.newLimitEuros}
							oninput={calculateTotalBudgetLimit}
						/>
					</td>
					<td>
						{#if salaryMismatch !== 0}
							<button class="secondary" onclick={() => fixBudget(budget)}>Fix</button>
						{/if}
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
{:else if user && budgets.length > 0 && wallets.length === 0}
	<p>No wallets found. Please create a wallet to generate transactions.</p>
{:else}
	<p>Loading of user data or budgets failed.</p>
{/if}

<button
	onclick={generateSalaryTransactions}
	disabled={salaryMismatch !== 0 ||
		selectedWalletId === undefined ||
		!transactionDescription.trim()}
>
	Generate Salary Transactions
</button>

<style>
	table {
		width: 100%;
		margin-bottom: 1rem;
	}
	th,
	td {
		text-align: left;
		padding: 0.5rem;
	}
	td:last-child {
		text-align: right;
		white-space: nowrap;
	}
	input,
	select {
		margin-bottom: 0;
	}
	.warning {
		background-color: var(--pico-form-element-background-color);
		color: var(--pico-color-red-600);
		border: 1px solid var(--pico-color-red-300);
		padding: 1rem;
		margin-bottom: 1rem;
		border-radius: var(--pico-border-radius);
	}
	button.secondary {
		background-color: var(--pico-secondary);
		color: var(--pico-secondary-inverse);
	}
	div {
		margin-bottom: 1rem;
	}
</style>
