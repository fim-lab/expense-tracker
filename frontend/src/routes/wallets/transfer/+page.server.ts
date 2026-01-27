import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
	const API_URL = '/api';
	const cookieHeader = cookies
		.getAll()
		.map((c) => `${c.name}=${c.value}`)
		.join('; ');

	const walletsRes = await fetch(`${API_URL}/wallets`, { headers: { Cookie: cookieHeader } });

	if (walletsRes.status === 401) throw error(401, 'Unauthorized');

	return {
		wallets: walletsRes.ok ? await walletsRes.json() : []
	};
};
