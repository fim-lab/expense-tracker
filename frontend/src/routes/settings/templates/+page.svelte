<script lang="ts">
	import { formatCurrency } from '$lib/utils';
	import type { Budget, TemplateGroup, TransactionTemplate, Wallet } from '$lib/types';

	let { data } = $props();

	let templates = $state<TransactionTemplate[]>(
		(data.templates || []).map((t: TransactionTemplate) => ({
			...t,
			newGroupId: t.groupId ?? undefined
		}))
	);

	let templateGroups = $state<TemplateGroup[]>(
		(data.templateGroups || []).map((g: TemplateGroup) => ({ ...g, isEditing: false, newName: '' }))
	);

	let wallets = $state<Wallet[]>(data.wallets || []);
	let budgets = $state<Budget[]>(data.budgets || []);

	let newGroupName = $state('');

	function walletName(walletId: number) {
		return wallets.find((w) => w.id === walletId)?.name || 'Unknown';
	}

	function budgetName(budgetId: number | null | undefined) {
		return budgets.find((b) => b.id === budgetId)?.name || 'None';
	}

	async function saveTemplateGroup(template: TransactionTemplate) {
		const res = await fetch(`/api/transaction-templates/${template.id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				...template,
				groupId: template.newGroupId ?? null
			})
		});

		if (res.ok) {
			template.groupId = template.newGroupId ?? null;
		} else {
			console.error('Failed to update transaction template');
			alert('Failed to update transaction template');
		}
	}

	function startEditingGroup(group: TemplateGroup) {
		group.isEditing = true;
		group.newName = group.name;
	}

	function cancelEditingGroup(group: TemplateGroup) {
		group.isEditing = false;
	}

	async function updateGroup(group: TemplateGroup) {
		if (!group.newName) {
			alert('Enter a name.');
			return;
		}
		const res = await fetch(`/api/template-groups/${group.id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ ...group, name: group.newName })
		});

		if (res.ok) {
			group.name = group.newName;
			group.isEditing = false;
		} else {
			console.error('Failed to update template group');
		}
	}

	async function deleteGroup(groupId: number) {
		if (confirm('Delete this template group? Templates in it will become ungrouped.')) {
			const res = await fetch(`/api/template-groups/${groupId}`, {
				method: 'DELETE'
			});

			if (res.ok) {
				templateGroups = templateGroups.filter((g) => g.id !== groupId);
				for (const template of templates) {
					if (template.groupId === groupId) {
						template.groupId = null;
						template.newGroupId = undefined;
					}
				}
			} else {
				console.error('Failed to delete template group');
			}
		}
	}

	async function createGroup() {
		if (!newGroupName) {
			alert('Enter a name.');
			return;
		}
		const res = await fetch('/api/template-groups', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ name: newGroupName })
		});

		if (res.ok) {
			const created = await res.json();
			templateGroups = [...templateGroups, { ...created, isEditing: false, newName: '' }];
			newGroupName = '';
		} else {
			console.error('Failed to create template group');
			alert('Failed to create template group');
		}
	}
</script>

<h1>Templates</h1>

<h2>Transaction Templates</h2>

{#if templates.length > 0}
	<table>
		<thead>
			<tr>
				<th>Description</th>
				<th>Day</th>
				<th>Amount</th>
				<th>Wallet</th>
				<th>Budget</th>
				<th>Group</th>
				<th>Actions</th>
			</tr>
		</thead>
		<tbody>
			{#each templates as template (template.id)}
				<tr>
					<td>{template.description}</td>
					<td>{template.day}</td>
					<td>{formatCurrency(template.amountInCents)}</td>
					<td>{walletName(template.walletId)}</td>
					<td>{budgetName(template.budgetId)}</td>
					<td>
						<select bind:value={template.newGroupId}>
							<option value={undefined}>Ungrouped</option>
							{#each templateGroups as group (group.id)}
								<option value={group.id}>{group.name}</option>
							{/each}
						</select>
					</td>
					<td>
						<button onclick={() => saveTemplateGroup(template)}>Save</button>
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
{:else}
	<p>No transaction templates found.</p>
{/if}

<h2>Template Groups</h2>

{#if templateGroups.length > 0}
	<table>
		<thead>
			<tr>
				<th>Name</th>
				<th>Actions</th>
			</tr>
		</thead>
		<tbody>
			{#each templateGroups as group (group.id)}
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
	<p>No template groups found.</p>
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
	input,
	select {
		margin-bottom: 0;
	}
</style>
