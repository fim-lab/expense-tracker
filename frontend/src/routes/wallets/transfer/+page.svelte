<script lang="ts">
	import { goto } from '$app/navigation';

	let { data } = $props();

	let fromWalletId = $state(0);
	let toWalletId = $state(0);
	let amount = $state(0);
	let errorMessage = $state('');

	const availableToWallets = $derived(() => {
		if (fromWalletId === 0) {
			return [];
		}
		return data.wallets.filter((wallet) => wallet.id !== fromWalletId);
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

	async function handleSubmit(e: Event) {
		e.preventDefault();
		errorMessage = '';

		if (fromWalletId === 0 || toWalletId === 0) {
			errorMessage = 'Please select both a "from" and "to" wallet.';
			return;
		}

		if (amount <= 0) {
			errorMessage = 'Amount must be greater than zero.';
			return;
		}

		const payload = {
			fromWalletId: Number(fromWalletId),
			toWalletId: Number(toWalletId),
			amount: Math.round(amount * 100)
		};

		const res = await fetch('/api/transactions/transfer', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(payload)
		});

		if (res.ok) {
			goto('/');
		} else {
			const errorText = await res.text();
			console.error('Backend Error:', errorText);
			errorMessage = `Failed to transfer money: ${errorText}`;
		}
	}
</script>

<article>
	<h3>Transfer Money Between Wallets</h3>
	<form onsubmit={handleSubmit}>
		<div class="grid">
			<label>
				From Wallet
				<select bind:value={fromWalletId} required>
					<option value={0} disabled>Select Wallet</option>
					{#each data.wallets as wallet}
						<option value={wallet.id}>{wallet.name} ({wallet.balanceCents / 100}€)</option>
					{/each}
				</select>
			</label>
			<label>
				To Wallet
				<select bind:value={toWalletId} required>
					<option value={0} disabled>Select Wallet</option>
					{#each availableToWallets() as wallet}
						<option value={wallet.id}>{wallet.name} ({wallet.balanceCents / 100}€)</option>
					{/each}
				</select>
			</label>
		</div>

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

		{#if errorMessage}
			<p class="error-message">{errorMessage}</p>
		{/if}

		<button type="submit">Transfer Money</button>
	</form>
</article>

<style>
	.error-message {
		color: var(--pico-del-color);
		margin-top: 1rem;
		margin-bottom: 0;
	}
</style>
