import type { PageServerLoad } from './$types';
import type { Wallet, Budget } from '$lib/types';

export const load: PageServerLoad<{ wallets: Wallet[]; budgets: Budget[] }> = async ({
	fetch
}) => {
	const [walletsRes, budgetsRes] = await Promise.all([
		fetch('/api/wallets'),
		fetch('/api/budgets')
	]);
	const wallets = walletsRes.ok ? ((await walletsRes.json()) as Wallet[]) : [];
	const budgets = budgetsRes.ok ? ((await budgetsRes.json()) as Budget[]) : [];

	return {
		wallets,
		budgets
	};
};
