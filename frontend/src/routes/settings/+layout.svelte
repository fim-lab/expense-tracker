<script lang="ts">
	let { children } = $props();
</script>

<div class="grid">
	<aside class="container">
		<nav>
			<ul>
				<li><a href="/settings/budgets">Budgets</a></li>
				<li><a href="/settings/wallets">Wallets</a></li>
				<li><a href="/settings/appearance">Appearance</a></li>
			</ul>
		</nav>
	</aside>
	<div>
		{@render children()}
	</div>
</div>

<style>
	aside {
        padding-right: 1rem;
		border-right: 1px solid var(--pico-muted-border-color);
	}
</style>