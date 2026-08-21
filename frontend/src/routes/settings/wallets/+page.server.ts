import type { PageServerLoad } from './$types';
import type { Depot, Portfolio, Stock } from '$lib/types';

export const load: PageServerLoad = async ({ fetch }) => {
	const [walletsRes, depotsRes, budgetsRes, stocksRes] = await Promise.all([
		fetch('/api/wallets'),
		fetch('/api/depots'),
		fetch('/api/budgets'),
		fetch('/api/stocks')
	]);

	const wallets = (walletsRes.ok ? await walletsRes.json() : []) ?? [];
	const depots: Depot[] = (depotsRes.ok ? await depotsRes.json() : []) ?? [];
	const budgets = (budgetsRes.ok ? await budgetsRes.json() : []) ?? [];
	const allStocks: Stock[] = (stocksRes.ok ? await stocksRes.json() : []) ?? [];

	const portfolioResponses = await Promise.all(
		depots.map((depot: Depot) => fetch(`/api/depots/${depot.id}/portfolio`))
	);
	const portfolios: Portfolio[] = await Promise.all(
		portfolioResponses.map((res) => (res.ok ? res.json() : { positions: [] }))
	);
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
