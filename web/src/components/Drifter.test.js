import { driftX, driftParams, hashUnit, OFFSCREEN_OFFSET } from './Drifter';

const WIDTH = 1000;
const SPAN = WIDTH + OFFSCREEN_OFFSET * 2;

describe('hashUnit', () => {
  test('is deterministic for the same input', () => {
    expect(hashUnit('default/api-1')).toBe(hashUnit('default/api-1'));
  });

  test('differs between inputs', () => {
    expect(hashUnit('default/api-1')).not.toBe(hashUnit('default/api-2'));
  });

  test('stays within [0, 1)', () => {
    for (const s of ['', 'a', 'default/very-long-pod-name-7d9f8', '🐱']) {
      const v = hashUnit(s);
      expect(v).toBeGreaterThanOrEqual(0);
      expect(v).toBeLessThan(1);
    }
  });
});

describe('driftX', () => {
  test('starts off-screen left at phase 0', () => {
    expect(driftX({ elapsedMs: 0, duration: 10, phase: 0, width: WIDTH })).toBe(-OFFSCREEN_OFFSET);
  });

  test('reaches off-screen right at the end of a cycle', () => {
    // Just before wrapping, to avoid landing exactly on the modulo boundary.
    const x = driftX({ elapsedMs: 9999, duration: 10, phase: 0, width: WIDTH });
    expect(x).toBeGreaterThan(WIDTH);
    expect(x).toBeLessThanOrEqual(-OFFSCREEN_OFFSET + SPAN);
  });

  test('wraps back to the left edge rather than drifting away forever', () => {
    const atStart = driftX({ elapsedMs: 0, duration: 10, phase: 0, width: WIDTH });
    const afterFive = driftX({ elapsedMs: 50000, duration: 10, phase: 0, width: WIDTH });
    expect(afterFive).toBeCloseTo(atStart, 6);
  });

  // The fix: a non-zero phase means the entity is already mid-flight at t=0.
  test('phase offsets the starting position', () => {
    const x = driftX({ elapsedMs: 0, duration: 10, phase: 0.5, width: WIDTH });
    expect(x).toBeCloseTo(-OFFSCREEN_OFFSET + SPAN * 0.5, 6);
  });

  test('stays within the track for any phase and time', () => {
    for (let i = 0; i < 200; i += 1) {
      const x = driftX({
        elapsedMs: i * 813,
        duration: 10 + (i % 30),
        phase: (i % 17) / 17,
        width: WIDTH,
      });
      expect(x).toBeGreaterThanOrEqual(-OFFSCREEN_OFFSET);
      expect(x).toBeLessThanOrEqual(-OFFSCREEN_OFFSET + SPAN);
    }
  });

  describe('mirror', () => {
    test('turns around at the far edge instead of wrapping', () => {
      const out = driftX({ elapsedMs: 5000, duration: 10, phase: 0, width: WIDTH, mirror: true });
      const back = driftX({ elapsedMs: 15000, duration: 10, phase: 0, width: WIDTH, mirror: true });
      // Same distance along the track, travelling in opposite directions.
      expect(back).toBeCloseTo(out, 6);
    });

    test('completes an out-and-back over two durations', () => {
      const start = driftX({ elapsedMs: 0, duration: 10, phase: 0, width: WIDTH, mirror: true });
      const returned = driftX({ elapsedMs: 20000, duration: 10, phase: 0, width: WIDTH, mirror: true });
      expect(returned).toBeCloseTo(start, 6);
    });
  });
});

describe('initial distribution', () => {
  // The actual regression being fixed: on load, entities must be spread along
  // the track instead of stacked at the left edge.
  const ids = Array.from({ length: 150 }, (_, i) => `ns-${i % 7}/pod-${i}`);
  const positionsAtLoad = ids.map((id) => {
    const { duration, phase } = driftParams(id);
    return driftX({ elapsedMs: 0, duration, phase, width: WIDTH });
  });

  test('does not stack everything at the spawn point', () => {
    const atSpawn = positionsAtLoad.filter((x) => x === -OFFSCREEN_OFFSET);
    expect(atSpawn.length).toBeLessThan(3);
  });

  test('covers the whole track, not just one end', () => {
    expect(Math.min(...positionsAtLoad)).toBeLessThan(-OFFSCREEN_OFFSET + SPAN * 0.1);
    expect(Math.max(...positionsAtLoad)).toBeGreaterThan(-OFFSCREEN_OFFSET + SPAN * 0.9);
  });

  test('spreads roughly evenly across the track', () => {
    const buckets = new Array(10).fill(0);
    positionsAtLoad.forEach((x) => {
      const i = Math.min(9, Math.floor(((x + OFFSCREEN_OFFSET) / SPAN) * 10));
      buckets[i] += 1;
    });
    // 150 entities over 10 buckets averages 15; allow wide slack for hashing
    // while still catching a swarm piled into one or two buckets.
    buckets.forEach((count) => expect(count).toBeGreaterThan(3));
  });
});

describe('stability across refetches', () => {
  // Params are name-derived, so a pod keeps its lane and speed when the list is
  // refreshed. Previously Math.random() re-rolled on every render.
  test('same id yields identical params', () => {
    expect(driftParams('default/api-7d9f')).toEqual(driftParams('default/api-7d9f'));
  });

  test('duration lands in the intended range', () => {
    for (let i = 0; i < 100; i += 1) {
      const { duration } = driftParams(`ns/pod-${i}`);
      expect(duration).toBeGreaterThanOrEqual(10);
      expect(duration).toBeLessThan(40);
    }
  });
});
