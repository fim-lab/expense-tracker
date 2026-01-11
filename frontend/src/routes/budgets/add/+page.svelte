<script lang="ts">
	import { goto } from '$app/navigation';

	let name = $state('');
	let limitInEuro = $state(0);
	let isSubmitting = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		isSubmitting = true;

		const payload = {
			id: 0, // Backend handles this
			userId: 0, // Backend handles via session
			name: name,
			limitCents: Math.round(limitInEuro * 100)
		};

		const res = await fetch('/api/budgets', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});

		if (res.ok) {
			goto('/');
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
			<input type="number" step="0.01" bind:value={limitInEuro} required />
		</label>

		<button type="submit" aria-busy={isSubmitting}>Create Budget</button>
	</form>
</article>
