<script lang="ts">
	import { formatCurrency } from '$lib/utils';
	import type { Budget, User } from '$lib/types';

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
			newName: '',
			newLimitEuros: b.limitCents / 100
		}))
	);

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
				budget.isEditing && budget.newLimitEuros !== undefined
					? Math.round(budget.newLimitEuros * 100)
					: budget.limitCents;
		}
		totalBudgetLimitCents = total;
		calculateSalaryMismatch();
	}

	function updateSalaryInput() {
		if (user && user.salaryCents !== undefined) {
			currentSalaryInput =
				user.isEditing && user.newSalaryEuros !== undefined
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

	function startEditing(budget: Budget) {
		budget.isEditing = true;
		budget.newName = budget.name;
		budget.newLimitEuros = budget.limitCents / 100;
	}

	function cancelEditing(budget: Budget) {
		budget.isEditing = false;
		calculateTotalBudgetLimit();
	}

	async function updateBudget(budget: Budget) {
		if (!budget.newName || !budget.newLimitEuros) {
			alert('Enter name or limit.');
			return;
		}
		const newLimitCents = Math.round(budget.newLimitEuros * 100);
		const res = await fetch(`/api/budgets/${budget.id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ ...budget, name: budget.newName, limitCents: newLimitCents })
		});

		if (res.ok) {
			budget.name = budget.newName;
			budget.limitCents = newLimitCents;
			budget.isEditing = false;
			calculateTotalBudgetLimit();
		} else {
			console.error('Failed to update budget');
		}
	}

	async function deleteBudget(budgetId: number) {
		if (confirm('Are you sure you want to delete this budget?')) {
			const res = await fetch(`/api/budgets/${budgetId}`, {
				method: 'DELETE'
			});

			if (res.ok) {
				budgets = budgets.filter((b) => b.id !== budgetId);
				calculateTotalBudgetLimit();
			} else {
				console.error('Failed to delete budget');
			}
		}
	}

	function startEditingSalary() {
		if (user) {
			user.isEditing = true;
			user.newSalaryEuros = user.salaryCents / 100;
		}
	}

	function cancelEditingSalary() {
		if (user) {
			user.isEditing = false;
			calculateSalaryMismatch();
		}
	}

	async function updateSalary() {
		if (!user || user.newSalaryEuros === undefined || user.newSalaryEuros < 0) {
			alert('Enter a valid salary.');
			return;
		}
		const newSalaryCents = Math.round(user.newSalaryEuros * 100);
		const res = await fetch(`/api/users/me/salary`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ salaryCents: newSalaryCents })
		});

		if (res.ok) {
			user.salaryCents = newSalaryCents;
			user.isEditing = false;
			calculateSalaryMismatch();
		} else {
			console.error('Failed to update salary');
			alert('Failed to update salary. Please try again.');
		}
	}
</script>

<h1>Salary & Budgets</h1>

{#if salaryMismatch !== 0}
	<article class="warning">
		The sum of your budget limits ({formatCurrency(totalBudgetLimitCents)}) does not equal your
		salary ({formatCurrency(currentSalaryInput)}).<br />
		The Difference is {formatCurrency(salaryMismatch)}
	</article>
{/if}

{#if user}
	<table>
		<thead>
			<tr>
				<th>Current Salary</th>
				<th>Actions</th>
			</tr>
		</thead>
		<tbody>
			<tr>
				<td>
					{#if user.isEditing}
						<input
							type="number"
							step="0.01"
							bind:value={user.newSalaryEuros}
							oninput={updateSalaryInput}
						/>
					{:else}
						{formatCurrency(user.salaryCents)}
					{/if}
				</td>
				<td>
					{#if user.isEditing}
						{#if salaryMismatch !== 0}
							<button class="secondary" onclick={fixSalary}>Fix</button>
						{/if}
						<button onclick={updateSalary}>Save</button>
						<button class="secondary" onclick={cancelEditingSalary}>Cancel</button>
					{:else}
						<button onclick={startEditingSalary}>Edit Salary</button>
					{/if}
				</td>
			</tr>
		</tbody>
	</table>
{:else}
	<p>Loading user data or user not found.</p>
{/if}

<h2>Budgets</h2>

{#if budgets.length > 0}
	<table>
		<thead>
			<tr>
				<th>Budget</th>
				<th>Limit</th>
				<th>Actions</th>
			</tr>
		</thead>
		<tbody>
			{#each budgets as budget (budget.id)}
				<tr>
					<td>
						{#if budget.isEditing}
							<input type="text" bind:value={budget.newName} />
						{:else}
							{budget.name}
						{/if}
					</td>
					<td>
						{#if budget.isEditing}
							<input
								type="number"
								step="0.01"
								bind:value={budget.newLimitEuros}
								oninput={calculateTotalBudgetLimit}
							/>
						{:else}
							{formatCurrency(budget.limitCents)}
						{/if}
					</td>
					<td>
						{#if budget.isEditing}
							{#if salaryMismatch !== 0}
								<button class="secondary" onclick={() => fixBudget(budget)}>Fix</button>
							{/if}
							<button onclick={() => updateBudget(budget)}>OK</button>
							<button class="secondary" onclick={() => cancelEditing(budget)}>Cancel</button>
						{:else}
							<button onclick={() => startEditing(budget)}>Edit</button>
							<span
								title={!budget.canDelete
									? 'Only budgets with a balance of 0 and no transactions can be deleted.'
									: ''}
							>
								<button
									class="secondary"
									onclick={() => deleteBudget(budget.id)}
									disabled={!budget.canDelete}>Delete</button
								>
							</span>
						{/if}
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
{:else}
	<p>No budgets found.</p>
{/if}

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
	input {
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
</style>
