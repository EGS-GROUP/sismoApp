import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import FeedFilters from './FeedFilters.svelte';

describe('FeedFilters', () => {
	it('renders all filter buttons', () => {
		render(FeedFilters, { props: { feedFilter: 'all' } });
		expect(screen.getByRole('button', { name: /todos/i })).toBeInTheDocument();
		expect(screen.getByRole('button', { name: /FUNVISIS/i })).toBeInTheDocument();
	});

	it('calls onChange when a filter is clicked', async () => {
		const onChange = vi.fn();
		render(FeedFilters, { props: { feedFilter: 'all', onChange } });
		const btn = screen.getByRole('button', { name: /FUNVISIS/i });
		await fireEvent.click(btn);
		expect(onChange).toHaveBeenCalledWith('FUNVISIS');
	});
});
