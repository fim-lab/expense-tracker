<script lang="ts">
	import { goto } from '$app/navigation';

	let name = $state('');
	let initialBalanceEuro = $state(0);
	let isSubmitting = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		isSubmitting = true;

		const payload = {
			id: 0,
			userId: 0,
			name: name,
			balance: Math.round(initialBalanceEuro * 100)
		};

		const res = await fetch('/api/wallets', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});

		if (res.ok) {
			goto('/');
		} else {
			alert('Failed to save wallet');
			isSubmitting = false;
		}
	}
</script>

<article>
	<header><strong>Add New Wallet</strong></header>
	<form onsubmit={handleSubmit}>
		<label>
			Wallet Name
			<input type="text" bind:value={name} placeholder="e.g. Cash, Revolut, Sparkasse" required />
		</label>

		<label>
			Current Balance (EUR)
			<input type="number" step="0.01" bind:value={initialBalanceEuro} required />
		</label>

		<button type="submit" aria-busy={isSubmitting}>Create Wallet</button>
	</form>
</article>
