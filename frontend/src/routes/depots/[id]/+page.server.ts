import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { Depot, Portfolio, TradeDTO, Wallet, Budget } from '$lib/types';

export const load: PageServerLoad = async ({ fetch, cookies, params }) => {
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

		if (res.status === 404) {
			throw error(404, 'Depot not found');
		}

		if (!res.ok) return null;
		return res.json();
	};

	const depot = (await authedApiFetch(`/depots/${params.id}`)) as Depot | null;
	if (!depot) {
		throw error(404, 'Depot not found');
	}

	const portfolio = ((await authedApiFetch(`/depots/${params.id}/portfolio`)) as Portfolio) ?? {
		depotId: depot.id,
		positions: [],
		investedInCents: 0,
		realizedGainInCents: 0
	};
	const trades = ((await authedApiFetch(`/depots/${params.id}/trades`)) as TradeDTO[]) ?? [];
	const wallets = ((await authedApiFetch('/wallets')) as Wallet[]) ?? [];
	const budgets = ((await authedApiFetch('/budgets')) as Budget[]) ?? [];

	return {
		depot,
		portfolio,
		trades,
		wallet: wallets.find((w) => w.id === depot.walletId),
		budget: budgets.find((b) => b.id === depot.budgetId)
	};
};
