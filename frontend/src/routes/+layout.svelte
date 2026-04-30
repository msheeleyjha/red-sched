<script lang="ts">
	import '../app.css';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import type { LayoutData } from './$types';

	export let data: LayoutData;

	$: user = data.user;

	const pendingAllowedPaths = ['/pending', '/referee/profile', '/auth'];

	$: {
		const path = $page.url.pathname;
		if (user) {
			if (path === '/') {
				if (user.role === 'pending_referee') {
					goto('/pending');
				} else {
					goto('/dashboard');
				}
			} else if (user.role === 'pending_referee' && !pendingAllowedPaths.some(p => path.startsWith(p))) {
				goto('/pending');
			} else if (path.startsWith('/admin') && user.role !== 'assignor') {
				goto('/dashboard');
			}
		}
	}
</script>

<slot />
