<script lang="ts">
	import type { PageProps } from './$types';
	import Alert from '$lib/components/Alert.svelte';
	import Button from '$lib/components/Button.svelte';
	import FormField from '$lib/components/FormField.svelte';

	let { form }: PageProps = $props();

	let submitting = $state(false);
</script>

<svelte:head>
	<title>Log in - starterapp</title>
</svelte:head>

<section class="login">
	<h1>Log in</h1>
	<form method="POST" onsubmit={() => (submitting = true)}>
		<FormField
			label="Email"
			name="email"
			type="email"
			value={form?.email ?? ''}
			error={form?.fieldErrors?.email}
			required
			autocomplete="username"
		/>
		<FormField
			label="Password"
			name="password"
			type="password"
			error={form?.fieldErrors?.password}
			required
			autocomplete="current-password"
		/>
		{#if form?.error}
			<Alert>{form.error}</Alert>
		{/if}
		<Button type="submit" loading={submitting}>Log in</Button>
	</form>
</section>

<style>
	.login {
		max-width: 20rem;
		margin: 3rem auto;
	}
	form {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}
</style>
