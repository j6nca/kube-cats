import React from 'react';
import { hashUnit } from './Drifter';
import './Starfield.css';

// Fixed, deterministic particles; avoids React re-renders and layout work.
const layers = [
  { count: 42, depth: 'far' },
  { count: 28, depth: 'mid' },
  { count: 16, depth: 'near' },
];

export const createStars = (count, depth) =>
  Array.from({ length: count }, (_, i) => ({
    id: `${depth}-${i}`,
    x: hashUnit(`${depth}:${i}:x`) * 100,
    y: hashUnit(`${depth}:${i}:y`) * 100,
    size: 1 + hashUnit(`${depth}:${i}:size`) * 1.5,
    delay: -hashUnit(`${depth}:${i}:phase`) * 90,
    opacity: 0.28 + hashUnit(`${depth}:${i}:opacity`) * 0.55,
  }));

const Starfield = () => (
  <div className="space-background" aria-hidden="true" data-testid="starfield">
    <div className="space-background__glow" />
    {layers.map(({ count, depth }) => (
      <div key={depth} className={`star-layer star-layer--${depth}`}>
        {createStars(count, depth).map((star) => (
          <span
            key={star.id}
            className="star"
            style={{
              left: `${star.x}%`,
              top: `${star.y}%`,
              width: star.size,
              height: star.size,
              opacity: star.opacity,
              animationDelay: `${star.delay}s`,
            }}
          />
        ))}
      </div>
    ))}
  </div>
);

export default React.memo(Starfield);
