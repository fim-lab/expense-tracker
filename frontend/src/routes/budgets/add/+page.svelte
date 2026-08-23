<script lang="ts">
	import { goto } from '$app/navigation';
	import type { BudgetGroup } from '$lib/types';

	let { data } = $props();

	let budgetGroups: BudgetGroup[] = data.budgetGroups || [];

	let name = $state('');
	let limitInEuro = $state(0);
	let groupId: number | undefined = $state(undefined);
	let isSubmitting = $state(false);

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
			// Also update the bound state immediately
			limitInEuro = 0;
		}
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		isSubmitting = true;

		const payload = {
			id: 0,
			userId: 0,
			name: name,
			limitCents: Math.round(limitInEuro * 100),
			groupId: groupId ?? null
		};

		const res = await fetch('/api/budgets', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});

		if (res.ok) {
			goto('/settings/budgets');
		} else {
			alert('Failed to save budget');
			isSubmitting = false;
		}
	}
</script>

<article>
	<header><strong>Add New Budget Category</strong></header>
	<form onsubmit={handleSubmit}>
		<label>
			Category Name
			<input type="text" bind:value={name} placeholder="e.g. Groceries" required />
		</label>

		<label>
			Monthly Limit (EUR)
			<input
				type="number"
				step="0.01"
				bind:value={limitInEuro}
				required
				onfocus={handleFocus}
				onblur={handleBlur}
			/>
		</label>

		<label>
			Group
			<select bind:value={groupId}>
				<option value={undefined}>Ungrouped</option>
				{#each budgetGroups as group (group.id)}
					<option value={group.id}>{group.name}</option>
				{/each}
			</select>
		</label>

		<button type="submit" aria-busy={isSubmitting}>Create Budget</button>
	</form>
</article>
