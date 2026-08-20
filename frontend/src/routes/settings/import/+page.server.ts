import { error } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	if (env.SHOW_IMPORT_SETTING !== 'true') {
		throw error(404, 'Not found');
	}
};
