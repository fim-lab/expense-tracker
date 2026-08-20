import { env } from '$env/dynamic/private';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async () => {
	return {
		showImport: env.SHOW_IMPORT_SETTING === 'true'
	};
};
