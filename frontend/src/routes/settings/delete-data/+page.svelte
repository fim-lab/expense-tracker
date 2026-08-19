<script lang="ts">
	import { goto } from '$app/navigation';

	let deleteConfirm = $state('');
	let isSubmitting = $state(false);
	let error = $state<string | null>(null);

	async function handleDelete(e: Event) {
		e.preventDefault();
		if (deleteConfirm !== 'delete') {
			return;
		}

		if (
			!confirm(
				'Are you ABSOLUTELY sure? This will delete all your transactions, budgets, wallets and depots. This action cannot be undone.'
			)
		) {
			return;
		}

		isSubmitting = true;
		error = null;

		const res = await fetch('/api/users/me/data', {
			method: 'DELETE'
		});

		if (res.ok) {
			alert('All data has been deleted.');
			goto('/');
		} else {
			error = 'Failed to delete data. Please try again.';
			isSubmitting = false;
		}
	}
</script>

<h1>Delete All Data</h1>

<article>
	<p style="color: var(--pico-color-red-600);">
		<strong>Warning:</strong> This action is permanent and cannot be undone. All your transactions, budgets,
		wallets, depots, stocks and templates will be deleted.
	</p>

	<form onsubmit={handleDelete}>
		<label>
			Type <strong>delete</strong> to confirm:
			<input
				type="text"
				bind:value={deleteConfirm}
				placeholder="delete"
				required
				autocomplete="off"
			/>
		</label>

		<button
			type="submit"
			class="danger"
			disabled={deleteConfirm !== 'delete' || isSubmitting}
			aria-busy={isSubmitting}
		>
			Delete Everything
		</button>
	</form>

	{#if error}
		<p style="color: var(--pico-color-red-600);">{error}</p>
	{/if}
</article>

<style>
	.danger {
		background-color: rgb(187, 20, 20);
		border-color: red;
	}
</style>
