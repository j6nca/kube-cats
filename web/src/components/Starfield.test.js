import { render, screen } from '@testing-library/react';
import Starfield, { createStars } from './Starfield';

test('star positions are deterministic and remain inside the viewport', () => {
  expect(createStars(12, 'far')).toEqual(createStars(12, 'far'));
  createStars(12, 'far').forEach(({ x, y, opacity }) => {
    expect(x).toBeGreaterThanOrEqual(0);
    expect(x).toBeLessThan(100);
    expect(y).toBeGreaterThanOrEqual(0);
    expect(y).toBeLessThan(100);
    expect(opacity).toBeGreaterThan(0);
    expect(opacity).toBeLessThanOrEqual(1);
  });
});

test('renders decorative, non-interactive stars across three layers', () => {
  const { container } = render(<Starfield />);
  expect(screen.getByTestId('starfield')).toHaveAttribute('aria-hidden', 'true');
  expect(container.querySelectorAll('.star-layer')).toHaveLength(3);
  expect(container.querySelectorAll('.star')).toHaveLength(86);
});
