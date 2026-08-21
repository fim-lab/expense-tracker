import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { Depot, Portfolio, Stock } from '$lib/types';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
	const cookieHeader = cookies
		.getAll()
		.map((c) => `${c.name}=${c.value}`)
		.join('; ');

	const authedApiFetch = async (path: string) => {
		const res = await fetch(`/api${path}`, {
			headers: { Cookie: cookieHeader }
		});

		if (res.status === 401) {
			throw redirect(302, '/login');
		}

		if (!res.ok) return null;
		return res.json();
	};

	const wallets = (await authedApiFetch('/wallets')) ?? [];
	const depots: Depot[] = (await authedApiFetch('/depots')) ?? [];
	const budgets = (await authedApiFetch('/budgets')) ?? [];
	const allStocks: Stock[] = (await authedApiFetch('/stocks')) ?? [];

	const portfolios: Portfolio[] = [];
	for (const depot of depots) {
		portfolios.push((await authedApiFetch(`/depots/${depot.id}/portfolio`)) ?? { positions: [] });
	}
	const heldStockIds = new Set(
		portfolios.flatMap((portfolio) => portfolio.positions.map((p) => p.stockId))
	);
	const stocks = allStocks.filter((stock: Stock) => heldStockIds.has(stock.id));

	return {
		wallets,
		depots,
		budgets,
		stocks
	};
};
