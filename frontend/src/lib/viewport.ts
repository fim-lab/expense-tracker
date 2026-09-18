export function isDesktopViewport(breakpointPx = 768): boolean {
	return typeof window === 'undefined' || window.innerWidth > breakpointPx;
}
