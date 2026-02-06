<script lang="ts">
	import { goto } from '$app/navigation';

	let username = '';
	let password = '';
	let error = '';

	async function handleSubmit() {
		error = '';

		const payload = {
			username: username,
			password: password
		};

		try {
			const response = await fetch('/auth/login', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(payload)
			});

			if (response.ok) {
				goto('/');
			} else {
				error = 'Invalid credentials';
			}
		} catch (e) {
			error = 'Connection failed';
		}
	}
</script>

<div class="container">
	<article>
		<header><strong>Login</strong></header>

		<form on:submit|preventDefault={handleSubmit}>
			<label>
				Username
				<input type="text" name="username" placeholder="admin" bind:value={username} required />
			</label>

			<label>
				Password
				<input
					type="password"
					name="password"
					placeholder="password"
					bind:value={password}
					required
				/>
			</label>

			{#if error}
				<p style="color: var(--pico-del-color);">{error}</p>
			{/if}

			<button type="submit" aria-busy={false}>Log in</button>
		</form>
	</article>
</div>

<style>
	.container {
		display: grid;
		place-items: center;
		height: 100vh;
	}
	article {
		width: 100%;
		max-width: 400px;
	}
</style>
