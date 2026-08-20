import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { Depot, Wallet, Budget } from '$lib/types';

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

	const depots = ((await authedApiFetch('/depots')) as Depot[]) ?? [];
	const wallets = ((await authedApiFetch('/wallets')) as Wallet[]) ?? [];
	const budgets = ((await authedApiFetch('/budgets')) as Budget[]) ?? [];

	return { depots, wallets, budgets };
};
