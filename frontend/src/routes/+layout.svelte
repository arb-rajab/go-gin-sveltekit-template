<script lang="ts">
	import type { Snippet } from 'svelte';
	import { resolve } from '$app/paths';
	import '../app.css';

	let { children }: { children: Snippet } = $props();
</script>

<a class="skip-link" href="#main">Skip to content</a>
<header class="site-header">
	<a class="brand" href={resolve('/')}>starterapp</a>
</header>
<main id="main">
	{@render children()}
</main>

<style>
	.skip-link {
		position: absolute;
		left: -999px;
	}
	.skip-link:focus {
		left: 0.5rem;
		top: 0.5rem;
	}
	.site-header {
		padding: 0.75rem 1.5rem;
		border-bottom: 1px solid var(--color-border);
	}
	.brand {
		font-weight: 600;
		color: var(--color-text);
		text-decoration: none;
	}
	main {
		max-width: 48rem;
		margin: 2rem auto;
		padding: 0 1rem;
	}
</style>
