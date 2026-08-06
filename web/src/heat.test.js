import { isHot, percent, podTooltip } from './heat';

const pressure = (cpu, memory, source = 'metrics') => ({ cpu, memory, source });

describe('isHot', () => {
  test('catches fire when cpu crosses the threshold', () => {
    expect(isHot(pressure(0.9, 0.1), 0.8)).toBe(true);
  });

  test('catches fire when memory crosses the threshold', () => {
    expect(isHot(pressure(0.1, 0.85), 0.8)).toBe(true);
  });

  test('stays cool below the threshold on both resources', () => {
    expect(isHot(pressure(0.5, 0.79), 0.8)).toBe(false);
  });

  test('treats the threshold as inclusive', () => {
    expect(isHot(pressure(0.8, 0), 0.8)).toBe(true);
  });

  // An unmeasurable pod reads as 0% pressure, which would otherwise look
  // identical to a genuinely idle one.
  test('never catches fire when the source is unknown', () => {
    expect(isHot({ cpu: 5, memory: 5, source: 'unknown' }, 0.8)).toBe(false);
  });

  test('tolerates missing pressure data', () => {
    expect(isHot(undefined, 0.8)).toBe(false);
    expect(isHot(null, 0.8)).toBe(false);
  });
});

describe('percent', () => {
  test('rounds to whole percentages', () => {
    expect(percent(0.847)).toBe('85%');
  });

  test('renders missing values as zero', () => {
    expect(percent(undefined)).toBe('0%');
  });
});

describe('podTooltip', () => {
  const pod = {
    name: 'api-7d9f',
    namespace: 'default',
    node: 'k-node-1',
    phase: 'Running',
    ready: true,
    restarts: 3,
    pressure: pressure(0.5, 0.25),
  };

  test('includes identity, health and pressure', () => {
    const text = podTooltip(pod);
    expect(text).toContain('api-7d9f');
    expect(text).toContain('namespace: default');
    expect(text).toContain('phase: Running (ready)');
    expect(text).toContain('restarts: 3');
    expect(text).toContain('cpu: 50%');
    expect(text).toContain('memory: 25%');
    expect(text).toContain('(from metrics)');
  });

  test('omits restarts when there are none', () => {
    expect(podTooltip({ ...pod, restarts: 0 })).not.toContain('restarts');
  });

  test('omits pressure when unmeasurable', () => {
    const text = podTooltip({ ...pod, pressure: { cpu: 0, memory: 0, source: 'unknown' } });
    expect(text).not.toContain('cpu:');
  });

  test('labels an unscheduled pod', () => {
    expect(podTooltip({ ...pod, node: '' })).toContain('node: unscheduled');
  });
});
