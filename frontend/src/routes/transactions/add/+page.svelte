<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import TransactionTemplateCard from '$lib/components/TransactionTemplateCard.svelte';
	import type { TransactionTemplate } from '$lib/types';
	let { data } = $props();

	const urlParams = page.url.searchParams;

	let description = $state(urlParams.get('description') || '');
	let amount = $state(Number(urlParams.get('amount')) || 0);
	let date = $state(new Date().toISOString().split('T')[0]);
	let walletId = $state(Number(urlParams.get('walletId')) || 0);
	let budgetId = $state(Number(urlParams.get('budgetId')) || 0);
	let type = $state(urlParams.get('type') || 'EXPENSE');
	let errorMessage = $state('');

	let templates: TransactionTemplate[] = $state(data.templates);

	async function handleDelete(templateId: number) {
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
		newDate.setDate(template.day + 1);
		date = newDate.toISOString().split('T')[0];
		description = template.description;
		amount = template.amountInCents / 100;
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
			budgetId: Number(budgetId),
			type: type,
			isPending: false,
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
			<input type="text" bind:value={description} placeholder="Grocery shopping..." required />
		</label>

		<div class="grid">
			<label>
				Amount (EUR)
				<input type="number" step="0.01" bind:value={amount} required />
			</label>
			<label>
				Wallet
				<select bind:value={walletId} required>
					<option value={0} disabled>Select Wallet</option>
					{#each data.wallets as wallet}
						<option value={wallet.id}>{wallet.name}</option>
					{/each}
				</select>
			</label>
		</div>

		<label>
			Budget
			<select bind:value={budgetId} required>
				<option value={0} disabled>Select Budget Category</option>
				{#each data.budgets as budget}
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

{#if templates.length > 0}
	<article>
		<h3>Templates</h3>
		<div>
			{#each templates as template (template.id)}
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
</style>
