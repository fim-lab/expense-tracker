<script lang="ts">
	let { data, children } = $props();
</script>

<div class="grid">
	<aside class="container">
		<nav>
			<ul>
				<li><a href="/settings/budgets">Budgets</a></li>
				<li><a href="/settings/wallets">Wallets & Depots</a></li>
				{#if data.showImport}
					<li><a href="/settings/import">Import</a></li>
				{/if}
				<li><a href="/settings/appearance">Appearance</a></li>
				<li><a href="/settings/delete-data" style="color: var(--pico-color-red-600);">Delete Data</a></li>
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
