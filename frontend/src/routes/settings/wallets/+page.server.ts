import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { Stock } from '$lib/types';

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
	const depots = (await authedApiFetch('/depots')) ?? [];
	const budgets = (await authedApiFetch('/budgets')) ?? [];
	const stocks: Stock[] = (await authedApiFetch('/portfolio/owned-stocks')) ?? [];

	return {
		wallets,
		depots,
		budgets,
		stocks
	};
};
