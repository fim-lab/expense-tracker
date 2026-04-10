import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch }) => {
	const [walletsRes, depotsRes] = await Promise.all([
		fetch('/api/wallets'),
		fetch('/api/depots')
	]);

	const wallets = walletsRes.ok ? await walletsRes.json() : [];
	const depots = depotsRes.ok ? await depotsRes.json() : [];

	return {
		wallets,
		depots
	};
};
