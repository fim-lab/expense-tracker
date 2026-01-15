<script lang="ts">
	const props = $props();

	const page = $derived(props.page);
	const totalPages = $derived(props.totalPages);

	const delta = 1;

	function createPageUrl(p: number) {
		return `?page=${p}`;
	}

	function pages() {
		const range: (number | '…')[] = [];

		const start = Math.max(1, page - delta);
		const end = Math.min(totalPages, page + delta);

		if (start > 1) {
			range.push(1);
			if (start > 2) range.push('…');
		}

		for (let i = start; i <= end; i++) {
			range.push(i);
		}

		if (end < totalPages) {
			if (end < totalPages - 1) range.push('…');
			range.push(totalPages);
		}

		return range;
	}
</script>

<nav aria-label="Pagination">
	<ul>
		{#if page > 1}
			<li>
				<a href={createPageUrl(page - 1)} class="secondary"> ← Prev </a>
			</li>
		{:else}
			<li>
				<span aria-disabled="true"> ← Prev </span>
			</li>
		{/if}

		<!-- Page numbers -->
		{#each pages() as p}
			<li>
				{#if p === '…'}
					<span aria-hidden="true">…</span>
				{:else if p === page}
					<a href={createPageUrl(p)} aria-current="page" style="font-weight: 600;">
						{p}
					</a>
				{:else}
					<a href={createPageUrl(p)}>
						{p}
					</a>
				{/if}
			</li>
		{/each}

		{#if page < totalPages}
			<li>
				<a href={createPageUrl(page + 1)} class="secondary"> Next → </a>
			</li>
		{:else}
			<li>
				<span aria-disabled="true"> Next → </span>
			</li>
		{/if}
	</ul>
</nav>

<style>
	nav ul {
		margin: 0 auto;
	}
</style>
