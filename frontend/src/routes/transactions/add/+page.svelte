<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import TransactionTemplateCard from '$lib/components/TransactionTemplateCard.svelte';
	import type { TemplateGroup, TransactionTemplate } from '$lib/types';
	import { formatCurrency } from '$lib/utils';
	let { data } = $props();

	const urlParams = page.url.searchParams;

	let description = $state(urlParams.get('description') || '');
	let amount = $state(Number(urlParams.get('amount')) || 0);
	let date = $state(new Date().toISOString().split('T')[0]);
	let walletId = $state(Number(urlParams.get('walletId')) || 0);
	let budgetId = $state(Number(urlParams.get('budgetId')) || 0);
	let type = $state(urlParams.get('type') || 'EXPENSE');
	let isDebt = $state(false);
	let errorMessage = $state('');

	let templates: TransactionTemplate[] = $state(data.templates || []);
	let templateGroups: TemplateGroup[] = $state(
		(data.templateGroups || []).map((g: TemplateGroup) => ({ ...g, isEditing: false, newName: '' }))
	);
	let fireSummary = $state('');

	function sortByPosition(list: TransactionTemplate[]) {
		return [...list].sort((a, b) => a.position - b.position || a.id - b.id);
	}

	let groupedTemplates = $derived(
		templateGroups.map((group) => ({
			group,
			templates: sortByPosition(templates.filter((t) => t.groupId === group.id))
		}))
	);

	let ungroupedTemplates = $derived(sortByPosition(templates.filter((t) => !t.groupId)));

	let expandedGroups = $state<Record<number, boolean>>({});

	function toggleGroup(groupId: number) {
		expandedGroups[groupId] = !expandedGroups[groupId];
	}

	let newGroupName = $state('');

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
			expandedGroups[created.id] = true;
			newGroupName = '';
		} else {
			console.error('Failed to create template group');
			alert('Failed to create template group');
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
		if (!confirm('Delete this template group? Templates in it will become ungrouped.')) return;
		const res = await fetch(`/api/template-groups/${groupId}`, {
			method: 'DELETE'
		});

		if (res.ok) {
			templateGroups = templateGroups.filter((g) => g.id !== groupId);
			templates = templates.map((t) => (t.groupId === groupId ? { ...t, groupId: null } : t));
		} else {
			console.error('Failed to delete template group');
		}
	}

	let draggedTemplateId = $state<number | null>(null);

	function handleDragStart(event: DragEvent, template: TransactionTemplate) {
		draggedTemplateId = template.id;
		event.dataTransfer?.setData('text/plain', String(template.id));
		if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move';
	}

	function handleDragEnd() {
		draggedTemplateId = null;
	}

	function handleCardDragOver(event: DragEvent) {
		event.preventDefault();
	}

	function handleCardDrop(event: DragEvent, targetTemplate: TransactionTemplate) {
		event.preventDefault();
		event.stopPropagation();
		moveTemplate(targetTemplate.groupId ?? null, targetTemplate.id);
	}

	function handleContainerDrop(event: DragEvent, groupId: number | null) {
		event.preventDefault();
		moveTemplate(groupId, null);
	}

	async function persistTemplateOrder(template: TransactionTemplate) {
		const res = await fetch(`/api/transaction-templates/${template.id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(template)
		});
		if (!res.ok) {
			console.error('Failed to update transaction template order');
		}
	}

	function moveTemplate(targetGroupId: number | null, beforeTemplateId: number | null) {
		if (draggedTemplateId == null) return;
		const draggedId = draggedTemplateId;
		draggedTemplateId = null;
		if (draggedId === beforeTemplateId) return;

		const dragged = templates.find((t) => t.id === draggedId);
		if (!dragged) return;

		const sourceGroupId = dragged.groupId ?? null;

		const destList = sortByPosition(
			templates.filter((t) => (t.groupId ?? null) === targetGroupId && t.id !== draggedId)
		);

		let insertIndex = destList.length;
		if (beforeTemplateId != null) {
			const idx = destList.findIndex((t) => t.id === beforeTemplateId);
			if (idx !== -1) insertIndex = idx;
		}

		destList.splice(insertIndex, 0, dragged);

		const updates = new Map<number, { position: number; groupId: number | null }>();
		destList.forEach((t, index) => {
			updates.set(t.id, { position: index, groupId: targetGroupId });
		});

		if (sourceGroupId !== targetGroupId) {
			const sourceList = sortByPosition(
				templates.filter((t) => (t.groupId ?? null) === sourceGroupId && t.id !== draggedId)
			);
			sourceList.forEach((t, index) => {
				updates.set(t.id, { position: index, groupId: sourceGroupId });
			});
		}

		const changed: TransactionTemplate[] = [];
		templates = templates.map((t) => {
			const u = updates.get(t.id);
			if (!u || (t.position === u.position && (t.groupId ?? null) === u.groupId)) return t;
			const updated = { ...t, position: u.position, groupId: u.groupId };
			changed.push(updated);
			return updated;
		});

		for (const t of changed) {
			persistTemplateOrder(t);
		}
	}

	let draftAmounts = $state<Record<number, number>>({});

	function handleAmountChange(templateId: number, amountInCents: number) {
		draftAmounts[templateId] = amountInCents;
	}

	function amountForTemplate(t: TransactionTemplate) {
		return draftAmounts[t.id] ?? t.amountInCents;
	}

	function netSumInCents(groupTemplates: TransactionTemplate[]) {
		return groupTemplates.reduce(
			(sum, t) => sum + (t.type === 'INCOME' ? amountForTemplate(t) : -amountForTemplate(t)),
			0
		);
	}

	$effect(() => {
		if (isDebt) budgetId = 0;
	});

	function handleFocus(event: FocusEvent) {
		const input = event.target as HTMLInputElement;
		if (input.value === '0') {
			input.value = '';
		}
	}

	function handleBlur(event: FocusEvent) {
		const input = event.target as HTMLInputElement;
		if (input.value === '') {
			input.value = '0';
		}
	}

	async function handleDelete(templateId: number) {
		if (!confirm('Are you sure you want to delete this template?')) return;
		const res = await fetch(`/api/transaction-templates/${templateId}`, {
			method: 'DELETE'
		});

		if (res.ok) {
			templates = templates.filter((t) => t.id !== templateId);
		} else {
			alert('Failed to delete template');
		}
	}

	function handleUse(template: TransactionTemplate) {
		const newDate = new Date();
		newDate.setDate(template.day);
		date = newDate.toISOString().split('T')[0];
		description = template.description;
		amount = amountForTemplate(template) / 100;
		walletId = template.walletId;
		if (template.budgetId) {
			budgetId = template.budgetId;
		}
		type = template.type;
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		errorMessage = '';

		const rfc3339Date = new Date(date).toISOString();

		const payload = {
			date: rfc3339Date,
			description: description,
			amountInCents: Math.round(amount * 100),
			walletId: Number(walletId),
			budgetId: isDebt ? null : Number(budgetId),
			type: type,
			isPending: false,
			isDebt: isDebt,
			tags: []
		};

		if (payload.amountInCents <= 0) {
			errorMessage = 'Amount must be greater than zero.';
			return;
		}
		if (payload.walletId === 0) {
			errorMessage = 'Please select a wallet.';
			return;
		}
		if (!isDebt && payload.budgetId === 0) {
			errorMessage = 'Please select a budget.';
			return;
		}

		const res = await fetch('/api/transactions', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});

		if (res.ok) {
			goto('/');
		} else {
			const errorText = await res.text();
			console.error('Backend Error:', errorText);
			errorMessage = `Failed to save transaction: ${errorText}`;
		}
	}

	async function handleFireGroup(groupTemplates: TransactionTemplate[]) {
		fireSummary = '';
		let succeeded = 0;
		let failed = 0;
		for (const template of groupTemplates) {
			const newDate = new Date();
			newDate.setDate(template.day);
			const payload = {
				date: newDate.toISOString(),
				description: template.description,
				amountInCents: amountForTemplate(template),
				walletId: template.walletId,
				budgetId: template.budgetId ?? null,
				type: template.type,
				isPending: false,
				isDebt: false,
				tags: template.tags ?? []
			};
			const res = await fetch('/api/transactions', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload)
			});
			if (res.ok) {
				succeeded++;
			} else {
				failed++;
			}
		}
		fireSummary = `Created ${succeeded} transaction(s)${failed ? `, ${failed} failed` : ''}.`;
	}

	async function saveAsTemplate(e: Event) {
		e.preventDefault();
		errorMessage = '';

		const day = new Date(date).getDate();

		const payload = {
			day: day,
			description: description,
			amountInCents: Math.round(amount * 100),
			walletId: Number(walletId),
			budgetId: Number(budgetId),
			type: type,
			tags: []
		};

		if (payload.amountInCents <= 0) {
			errorMessage = 'Amount must be greater than zero.';
			return;
		}
		if (payload.walletId === 0 || payload.budgetId === 0) {
			errorMessage = 'Please select a wallet and a budget.';
			return;
		}

		const res = await fetch('/api/transaction-templates', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});

		if (res.ok) {
			const newTemplate: TransactionTemplate = await res.json();
			templates = [...templates, newTemplate];
		} else {
			const errorText = await res.text();
			console.error('Backend Error:', errorText);
			errorMessage = `Failed to save transaction template: ${errorText}`;
		}
	}
</script>

<article>
	<h3>Add New Transaction</h3>
	<form onsubmit={handleSubmit}>
		<div class="grid">
			<label>
				Date
				<input type="date" bind:value={date} required />
			</label>
			<label>
				Type
				<select bind:value={type}>
					<option value="EXPENSE">Expense</option>
					<option value="INCOME">Income</option>
				</select>
			</label>
		</div>

		<label>
			Description
			<span
				class="tooltip-info"
				title="Use '$date' in templates for current month(or last at the beginning of the month) in the form 'Februar 26'"
				>ⓘ</span
			>
			<input type="text" bind:value={description} placeholder="Grocery shopping..." required />
		</label>
		<div class="grid">
			<label>
				Amount (EUR)
				<input
					type="number"
					onfocus={handleFocus}
					onblur={handleBlur}
					step="0.01"
					bind:value={amount}
					required
				/>
			</label>
			<label>
				Wallet
				<select bind:value={walletId} required>
					<option value={0} disabled>Select Wallet</option>
					{#each data.wallets || [] as wallet}
						<option value={wallet.id}>{wallet.name}</option>
					{/each}
				</select>
			</label>
		</div>

		<label>
			<input type="checkbox" bind:checked={isDebt} />
			This is a debt transaction
		</label>

		<label>
			Budget
			<select bind:value={budgetId} disabled={isDebt} required={!isDebt}>
				<option value={0} disabled>Select Budget Category</option>
				{#each data.budgets || [] as budget}
					<option value={budget.id}>{budget.name}</option>
				{/each}
			</select>
		</label>

		{#if errorMessage}
			<p class="error-message">{errorMessage}</p>
		{/if}

		<div class="grid">
			<button type="submit">Save Transaction</button>
			<button type="button" onclick={saveAsTemplate}>Save as template</button>
		</div>
	</form>
</article>

{#if fireSummary}
	<p>{fireSummary}</p>
{/if}

<div class="new-group-row">
	<input type="text" placeholder="New group name" bind:value={newGroupName} />
	<button type="button" class="new-group-btn" onclick={createGroup}>Add group</button>
</div>

{#each groupedTemplates as entry (entry.group.id)}
	<article
		class="group-article"
		ondragover={(e) => e.preventDefault()}
		ondrop={(e) => handleContainerDrop(e, entry.group.id)}
	>
		<div class="group-header">
			{#if entry.group.isEditing}
				<input type="text" class="group-name-input" bind:value={entry.group.newName} />
				<div class="group-actions">
					<button type="button" onclick={() => updateGroup(entry.group)}>OK</button>
					<button type="button" class="secondary" onclick={() => cancelEditingGroup(entry.group)}>
						Cancel
					</button>
				</div>
			{:else}
				<button type="button" class="group-toggle" onclick={() => toggleGroup(entry.group.id)}>
					<span class="chevron">{expandedGroups[entry.group.id] ? '▼' : '▶'}</span>
					<span class="group-name">{entry.group.name}</span>
					<span class="group-meta">
						{entry.templates.length} template{entry.templates.length === 1 ? '' : 's'} · net {formatCurrency(
							netSumInCents(entry.templates)
						)}
					</span>
				</button>
				<div class="group-actions">
					<button
						type="button"
						class="secondary"
						onclick={() => startEditingGroup(entry.group)}
					>
						Rename
					</button>
					<button type="button" class="secondary" onclick={() => deleteGroup(entry.group.id)}>
						Delete
					</button>
					<button type="button" class="create-btn" onclick={() => handleFireGroup(entry.templates)}>
						Create
					</button>
				</div>
			{/if}
		</div>
		{#if expandedGroups[entry.group.id]}
			<div
				class="group-templates"
				role="list"
				ondragover={(e) => e.preventDefault()}
				ondrop={(e) => handleContainerDrop(e, entry.group.id)}
			>
				{#each entry.templates as template (template.id)}
					<TransactionTemplateCard
						{template}
						ondelete={handleDelete}
						onuse={handleUse}
						editable
						onamountchange={handleAmountChange}
						draggable={true}
						dragging={draggedTemplateId === template.id}
						ondragstart={handleDragStart}
						ondragend={handleDragEnd}
						ondragover={handleCardDragOver}
						ondrop={handleCardDrop}
					/>
				{/each}
				{#if entry.templates.length === 0}
					<p class="empty-drop-hint">Drag templates here.</p>
				{/if}
			</div>
		{/if}
	</article>
{/each}

<article
	class="group-article"
	ondragover={(e) => e.preventDefault()}
	ondrop={(e) => handleContainerDrop(e, null)}
>
	<h3>Ungrouped</h3>
	<div
		class="group-templates"
		role="list"
		ondragover={(e) => e.preventDefault()}
		ondrop={(e) => handleContainerDrop(e, null)}
	>
		{#each ungroupedTemplates as template (template.id)}
			<TransactionTemplateCard
				{template}
				ondelete={handleDelete}
				onuse={handleUse}
				draggable={true}
				dragging={draggedTemplateId === template.id}
				ondragstart={handleDragStart}
				ondragend={handleDragEnd}
				ondragover={handleCardDragOver}
				ondrop={handleCardDrop}
			/>
		{/each}
		{#if ungroupedTemplates.length === 0}
			<p class="empty-drop-hint">No ungrouped templates.</p>
		{/if}
	</div>
</article>

<style>
	.error-message {
		color: var(--pico-del-color);
		margin-top: 1rem;
		margin-bottom: 0;
	}

	.tooltip-info {
		cursor: help;
		font-weight: bold;
		color: var(--pico-secondary); /* Using a secondary color for visibility */
		margin-left: 0.5rem;
	}

	.group-article {
		padding: 1rem;
	}

	.group-header {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.group-toggle {
		flex: 1;
		display: flex;
		align-items: baseline;
		gap: 0.5rem;
		background: none;
		border: none;
		padding: 0;
		margin: 0;
		text-align: left;
		cursor: pointer;
		color: inherit;
		font: inherit;
	}

	.chevron {
		font-size: 0.75rem;
	}

	.group-name {
		font-weight: bold;
	}

	.group-meta {
		color: var(--pico-secondary);
		font-size: 0.9rem;
	}

	.create-btn {
		flex-shrink: 0;
		width: auto;
	}

	.group-templates {
		margin-top: 1rem;
		min-height: 2rem;
	}

	.new-group-row {
		display: flex;
		gap: 0.5rem;
		align-items: center;
		margin-bottom: 1rem;
	}

	.new-group-row input {
		margin: 0;
	}

	.new-group-btn {
		flex-shrink: 0;
		width: auto;
		white-space: nowrap;
	}

	.group-name-input {
		flex: 1;
		margin: 0;
	}

	.group-actions {
		display: flex;
		gap: 0.5rem;
		flex-shrink: 0;
	}

	.group-actions button {
		width: auto;
	}

	.empty-drop-hint {
		color: var(--pico-muted-color);
		font-size: 0.9rem;
		text-align: center;
		margin: 0;
		padding: 0.5rem;
		border: 1px dashed var(--pico-muted-border-color);
		border-radius: var(--pico-border-radius);
	}
</style>
