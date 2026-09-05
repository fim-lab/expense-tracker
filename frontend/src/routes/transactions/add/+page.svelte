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
	let templateGroups: TemplateGroup[] = $state(data.templateGroups || []);
	let fireSummary = $state('');

	let groupedTemplates = $derived(
		templateGroups
			.map((group) => ({
				group,
				templates: templates.filter((t) => t.groupId === group.id)
			}))
			.filter((entry) => entry.templates.length > 0)
	);

	let ungroupedTemplates = $derived(templates.filter((t) => !t.groupId));

	let expandedGroups = $state<Record<number, boolean>>({});

	function toggleGroup(groupId: number) {
		expandedGroups[groupId] = !expandedGroups[groupId];
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

{#each groupedTemplates as entry (entry.group.id)}
	<article class="group-article">
		<div class="group-header">
			<button type="button" class="group-toggle" onclick={() => toggleGroup(entry.group.id)}>
				<span class="chevron">{expandedGroups[entry.group.id] ? '▼' : '▶'}</span>
				<span class="group-name">{entry.group.name}</span>
				<span class="group-meta">
					{entry.templates.length} template{entry.templates.length === 1 ? '' : 's'} · net {formatCurrency(
						netSumInCents(entry.templates)
					)}
				</span>
			</button>
			<button type="button" class="create-btn" onclick={() => handleFireGroup(entry.templates)}>
				Create
			</button>
		</div>
		{#if expandedGroups[entry.group.id]}
			<div class="group-templates">
				{#each entry.templates as template (template.id)}
					<TransactionTemplateCard
						{template}
						ondelete={handleDelete}
						onuse={handleUse}
						editable
						onamountchange={handleAmountChange}
					/>
				{/each}
			</div>
		{/if}
	</article>
{/each}

{#if ungroupedTemplates.length > 0}
	<article>
		<h3>Templates</h3>
		<div>
			{#each ungroupedTemplates as template (template.id)}
				<TransactionTemplateCard {template} ondelete={handleDelete} onuse={handleUse} />
			{/each}
		</div>
	</article>
{/if}

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
		margin: 0 0 0 auto;
	}

	.group-templates {
		margin-top: 1rem;
	}
</style>
