import config from './config';

// The API reports pressure as a ratio per resource plus the source it was
// derived from. "unknown" means there was no sensible denominator (no limits,
// no allocatable capacity), so those entities never catch fire rather than
// reading as 0% and looking deceptively healthy.
export const isHot = (pressure, threshold = config.heat_threshold) => {
  if (!config.show_fire) return false;
  if (!pressure || pressure.source === 'unknown') return false;
  return pressure.cpu >= threshold || pressure.memory >= threshold;
};

export const percent = (value) => `${Math.round((value || 0) * 100)}%`;

// Detail shown on hover when SHOW_TOOLTIPS is on. Kept plain text because it
// rides on the native title attribute.
export const podTooltip = (pod) => {
  const lines = [
    pod.name,
    `namespace: ${pod.namespace}`,
    `node: ${pod.node || 'unscheduled'}`,
    `phase: ${pod.phase}${pod.ready ? ' (ready)' : ''}`,
  ];

  if (pod.restarts > 0) lines.push(`restarts: ${pod.restarts}`);
  if (pod.pressure && pod.pressure.source !== 'unknown') {
    lines.push(`cpu: ${percent(pod.pressure.cpu)}  memory: ${percent(pod.pressure.memory)}`);
    lines.push(`(from ${pod.pressure.source})`);
  }

  return lines.join('\n');
};

export const nodeTooltip = (node, podCount) => {
  const lines = [
    node.name,
    `pods: ${podCount}`,
    `ready: ${node.ready ? 'yes' : 'no'}`,
    `cpu: ${node.resources.cpuAllocatable} allocatable`,
    `memory: ${node.resources.memoryAllocatable} allocatable`,
  ];

  if (node.pressure && node.pressure.source !== 'unknown') {
    lines.push(`cpu: ${percent(node.pressure.cpu)}  memory: ${percent(node.pressure.memory)}`);
    lines.push(`(from ${node.pressure.source})`);
  }

  return lines.join('\n');
};
