<script lang="ts">
	import { formatCurrency } from '$lib/utils';
	import type { Budget } from '$lib/types';

	let { data } = $props();

	let budgets = $state<Budget[]>(
		data.budgets.map((b) => ({
			...b,
			isEditing: false,
			newName: '',
			newLimitEuros: b.limitCents / 100
		}))
	);

	function startEditing(budget: Budget) {
		budget.isEditing = true;
		budget.newName = budget.name;
		budget.newLimitEuros = budget.limitCents / 100;
	}

	function cancelEditing(budget: Budget) {
		budget.isEditing = false;
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
			} else {
				console.error('Failed to delete budget');
			}
		}
	}
</script>

<h1>Budgets</h1>

{#if budgets.length > 0}
	<table>
		<thead>
			<tr>
				<th>Name</th>
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
							<input type="number" step="0.01" bind:value={budget.newLimitEuros} />
						{:else}
							{formatCurrency(budget.limitCents)}
						{/if}
					</td>
					<td>
						{#if budget.isEditing}
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
</style>
