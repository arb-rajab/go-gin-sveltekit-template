import { createRawSnippet } from 'svelte';

/** A children/footer snippet that renders plain text, for component tests. */
export function textSnippet(text: string) {
	return createRawSnippet(() => ({ render: () => `<span>${text}</span>` }));
}
