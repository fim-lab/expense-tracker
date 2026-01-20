<script lang="ts">
	import { page } from '$app/state';
	let { ...props } = $props();

	const pageNr = $derived(Number(page.url.searchParams.get("page")));

	const delta = 1;

	const pageUrl = (p: number) => {
		const url = new URL(page.url);
		url.searchParams.set('page', p.toString());
		return url.toString();
	};

	function pages() {
		const range: (number | '…')[] = [];

		const start = Math.max(1, pageNr - delta);
		const end = Math.min(props.totalPages, pageNr + delta);

		if (start > 1) {
			range.push(1);
			if (start > 2) range.push('…');
		}

		for (let i = start; i <= end; i++) {
			range.push(i);
		}

		if (end < props.totalPages) {
			if (end < props.totalPages - 1) range.push('…');
			range.push(props.totalPages);
		}

		return range;
	}
</script>

<nav aria-label="Pagination">
	<ul>
		{#if pageNr > 1}
			<li>
				<a href={pageUrl(pageNr - 1)} class="secondary"> ← Prev </a>
			</li>
		{:else}
			<li>
				<span aria-disabled="true"> ← Prev </span>
			</li>
		{/if}

		{#each pages() as p}
			<li>
				{#if p === '…'}
					<span aria-hidden="true">…</span>
				{:else if p === pageNr}
					<a href={pageUrl(p)} aria-current="page" style="font-weight: 600;">
						{p}
					</a>
				{:else}
					<a href={pageUrl(p)}>
						{p}
					</a>
				{/if}
			</li>
		{/each}

		{#if pageNr < props.totalPages}
			<li>
				<a href={pageUrl(pageNr + 1)} class="secondary"> Next → </a>
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
