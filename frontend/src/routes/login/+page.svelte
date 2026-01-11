<script lang="ts">
	import { goto } from '$app/navigation';

	let username = '';
	let password = '';
	let error = '';

	async function handleSubmit() {
		// 1. Reset error
		error = '';

		// 2. Construct the JSON body strictly matching the Go struct
		const payload = {
			username: username,
			password: password
		};

		try {
			// 3. Send to Go Backend (via Proxy)
			const response = await fetch('/auth/login', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(payload)
			});

			if (response.ok) {
				// Success: The Go backend set the HttpOnly cookie.
				// We just need to move the user to the dashboard.
				goto('/');
			} else {
				// Handle non-200 responses (e.g. 401 Unauthorized)
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
