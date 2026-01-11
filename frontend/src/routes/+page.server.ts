import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
  // 1. Fetch Wallets (or a dedicated /me endpoint if you have one)
  // We use the SvelteKit `fetch` wrapper. It automatically carries over 
  // headers/cookies from the browser request if configured, but since 
  // we are in a container, we rely on the proxy setup.
  
  // However, specifically in +page.server.ts, we are making the request 
  // from the Node container to the Go container directly.
  
  // NOTE: Because we are inside the container network, we might need 
  // to hit the Go backend directly on its internal port, OR use the 
  // standard Caddy route. 
  
  // Let's use the local API URL for internal communication to avoid 
  // SSL/Host issues.
  const API_URL = 'http://localhost:8080/api';

  // We need to pass the "Cookie" header manually from the browser request 
  // to the backend request so Go knows who we are.
  const cookieHeader = cookies.getAll().map(c => `${c.name}=${c.value}`).join('; ');

  try {
    const res = await fetch(`${API_URL}/wallets`, {
        headers: {
            // Forward the auth cookie
            Cookie: cookieHeader
        }
    });

    if (res.status === 401) {
      // Go said "Unauthorized", so we kick the user out
      throw redirect(302, '/login');
    }

    if (!res.ok) {
        return { wallets: [], transactions: [] }; // Handle errors gracefully
    }

    const wallets = await res.json();

    // You can fetch transactions here in parallel as well
    const transRes = await fetch(`${API_URL}/transactions`, {
        headers: { Cookie: cookieHeader }
    });
    const transactions = transRes.ok ? await transRes.json() : [];

    // Return data to the .svelte page
    return {
      wallets,
      transactions
    };

  } catch (err) {
    // If it's a redirect, re-throw it so SvelteKit handles it
    if ((err as any).status === 302) throw err;

    console.error("Dashboard Load Error:", err);
    return { wallets: [], transactions: [] };
  }
};