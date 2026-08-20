import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch }) => {
	const [walletsRes, depotsRes, budgetsRes] = await Promise.all([
		fetch('/api/wallets'),
		fetch('/api/depots'),
		fetch('/api/budgets')
	]);

	const wallets = walletsRes.ok ? await walletsRes.json() : [];
	const depots = depotsRes.ok ? await depotsRes.json() : [];
	const budgets = budgetsRes.ok ? await budgetsRes.json() : [];

	return {
		wallets,
		depots,
		budgets
	};
};
