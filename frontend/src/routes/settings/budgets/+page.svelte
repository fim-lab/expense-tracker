<script lang="ts">
	import { formatCurrency } from '$lib/utils';
	import type { Budget, BudgetGroup } from '$lib/types';

	let { data } = $props();

	let budgets = $state<Budget[]>(
		(data.budgets || []).map((b: Budget) => ({
			...b,
			isEditing: false,
			newName: '',
			newLimitEuros: b.limitCents / 100,
			newGroupId: b.groupId ?? undefined
		}))
	);

	let budgetGroups = $state<BudgetGroup[]>(
		(data.budgetGroups || []).map((g: BudgetGroup) => ({ ...g, isEditing: false, newName: '' }))
	);

	let newGroupName = $state('');

	function groupName(groupId: number | null | undefined) {
		return budgetGroups.find((g) => g.id === groupId)?.name || 'Ungrouped';
	}

	function startEditingGroup(group: BudgetGroup) {
		group.isEditing = true;
		group.newName = group.name;
	}

	function cancelEditingGroup(group: BudgetGroup) {
		group.isEditing = false;
	}

	async function updateGroup(group: BudgetGroup) {
		if (!group.newName) {
			alert('Enter a name.');
			return;
		}
		const res = await fetch(`/api/budget-groups/${group.id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ ...group, name: group.newName })
		});

		if (res.ok) {
			group.name = group.newName;
			group.isEditing = false;
		} else {
			console.error('Failed to update budget group');
		}
	}

	async function deleteGroup(groupId: number) {
		if (confirm('Delete this budget group? Budgets in it will become ungrouped.')) {
			const res = await fetch(`/api/budget-groups/${groupId}`, {
				method: 'DELETE'
			});

			if (res.ok) {
				budgetGroups = budgetGroups.filter((g) => g.id !== groupId);
				for (const budget of budgets) {
					if (budget.groupId === groupId) {
						budget.groupId = null;
					}
				}
			} else {
				console.error('Failed to delete budget group');
			}
		}
	}

	async function createGroup() {
		if (!newGroupName) {
			alert('Enter a name.');
			return;
		}
		const res = await fetch('/api/budget-groups', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ name: newGroupName })
		});

		if (res.ok) {
			const created = await res.json();
			budgetGroups = [...budgetGroups, { ...created, isEditing: false, newName: '' }];
			newGroupName = '';
		} else {
			console.error('Failed to create budget group');
			alert('Failed to create budget group');
		}
	}

	function startEditing(budget: Budget) {
		budget.isEditing = true;
		budget.newName = budget.name;
		budget.newLimitEuros = budget.limitCents / 100;
		budget.newGroupId = budget.groupId ?? undefined;
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
			body: JSON.stringify({
				...budget,
				name: budget.newName,
				limitCents: newLimitCents,
				groupId: budget.newGroupId ?? null
			})
		});

		if (res.ok) {
			budget.name = budget.newName;
			budget.limitCents = newLimitCents;
			budget.groupId = budget.newGroupId ?? null;
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

<h2>Budgets</h2>

{#if budgets.length > 0}
	<table>
		<thead>
			<tr>
				<th>Budget</th>
				<th>Limit</th>
				<th>Group</th>
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
							<select bind:value={budget.newGroupId}>
								<option value={undefined}>Ungrouped</option>
								{#each budgetGroups as group (group.id)}
									<option value={group.id}>{group.name}</option>
								{/each}
							</select>
						{:else}
							{groupName(budget.groupId)}
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

<a href="/budgets/add" role="button">Add new Budget</a>

<h2>Budget Groups</h2>

{#if budgetGroups.length > 0}
	<table>
		<thead>
			<tr>
				<th>Name</th>
				<th>Actions</th>
			</tr>
		</thead>
		<tbody>
			{#each budgetGroups as group (group.id)}
				<tr>
					<td>
						{#if group.isEditing}
							<input type="text" bind:value={group.newName} />
						{:else}
							{group.name}
						{/if}
					</td>
					<td>
						{#if group.isEditing}
							<button onclick={() => updateGroup(group)}>OK</button>
							<button class="secondary" onclick={() => cancelEditingGroup(group)}>Cancel</button>
						{:else}
							<button onclick={() => startEditingGroup(group)}>Edit</button>
							<button class="secondary" onclick={() => deleteGroup(group.id)}>Delete</button>
						{/if}
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
{:else}
	<p>No budget groups found.</p>
{/if}

<table>
	<tbody>
		<tr>
			<td><input type="text" placeholder="Group name" bind:value={newGroupName} /></td>
			<td><button onclick={createGroup}>Add Group</button></td>
		</tr>
	</tbody>
</table>

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
	th:last-child,
	td:last-child {
		text-align: right;
		white-space: nowrap;
	}
	input {
		margin-bottom: 0;
	}
</style>
