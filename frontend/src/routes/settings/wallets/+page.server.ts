import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch }) => {
    const res = await fetch('/api/wallets');
    if (res.ok) {
        const wallets = await res.json();
        return { wallets };
    }
    return { wallets: [] };
};
