import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { TransactionTemplate } from '$lib/types';
import {
	getLastMonthShortMonthYearString,
	getNextMonthShortMonthYearString,
	getShortMonthYearString
} from '$lib/utils';

const DAY_OF_MONTH_UNTIL_LAST_MONTH_IS_USED = 10;

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

	const wallets = (await authedApiFetch('/wallets')) || [];
	const budgets = (await authedApiFetch('/budgets')) || [];
	const templates = (await authedApiFetch('/transaction-templates')) || [];
	const templateGroups = (await authedApiFetch('/template-groups')) || [];
	let lastMonthString: string;
	let nextMonthString: string;
	if (displayLastMonthString()) {
		lastMonthString = getLastMonthShortMonthYearString(new Date());
		nextMonthString = getShortMonthYearString(new Date());
	} else {
		lastMonthString = getShortMonthYearString(new Date());
		nextMonthString = getNextMonthShortMonthYearString(new Date());
	}
	templates.forEach((tt: TransactionTemplate) => {
		tt.description = tt.description
			.replace('$lastmonth', lastMonthString)
			.replace('$nextmonth', nextMonthString);
	});

	return {
		wallets,
		budgets,
		templates,
		templateGroups
	};

	function displayLastMonthString() {
		return new Date().getDate() <= DAY_OF_MONTH_UNTIL_LAST_MONTH_IS_USED;
	}
};
