import type { PageServerLoad } from './$types';
import type { Wallet } from '$lib/types';

export const load: PageServerLoad<{ wallets: Wallet[] }> = async ({ fetch }) => {
	const res = await fetch('/api/wallets');
	const wallets = res.ok ? ((await res.json()) as Wallet[]) : [];

	return {
		wallets
	};
};
