<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import Alert from '$lib/components/Alert.svelte';
</script>

<svelte:head>
	<title>{page.status} - starterapp</title>
</svelte:head>

<h1>{page.status}</h1>
<Alert>{page.error?.message ?? 'Something went wrong.'}</Alert>
<p><a href={resolve('/')}>Back to the start</a></p>
