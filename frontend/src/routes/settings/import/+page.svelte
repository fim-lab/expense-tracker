<script lang="ts">
	let files: FileList;
	let count = 0;
	let loading = false;

	async function upload() {
		if (!files || files.length === 0) {
			alert('Please select a file');
			return;
		}

		loading = true;
		const formData = new FormData();
		formData.append('file', files[0]);

		try {
			const response = await fetch('/api/transactions/import', {
				method: 'POST',
				body: formData
			});

			if (response.ok) {
				const result = await response.json();
				count = result.count;
			} else {
				alert('Error uploading file');
			}
		} catch (error) {
			console.error('Error uploading file:', error);
			alert('Error uploading file');
		} finally {
			loading = false;
		}
	}
</script>

<h1 class="text-2xl font-bold">Import Transactions</h1>

<form on:submit|preventDefault={upload}>
	<label for="file">JSON File</label>
	<input type="file" id="file" name="file" accept=".json" bind:files />
	<button type="submit" disabled={loading}>{loading ? 'Uploading...' : 'Upload'}</button>
</form>

{#if count > 0}
	<p>Successfully imported {count} transactions.</p>
{/if}
