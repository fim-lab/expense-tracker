import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch }) => {
	const [depotsRes, walletsRes] = await Promise.all([
		fetch('/api/depots'),
		fetch('/api/wallets')
	]);

	const depots = depotsRes.ok ? await depotsRes.json() : [];
	const wallets = walletsRes.ok ? await walletsRes.json() : [];

	return {
		depots,
		wallets
	};
};
