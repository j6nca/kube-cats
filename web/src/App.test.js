import { render, screen, waitFor } from '@testing-library/react';
import App from './App';

const node = (name, pressure) => ({
  name,
  ready: true,
  resources: { cpuAllocatable: '4', memoryAllocatable: '8Gi' },
  pressure,
});

const pod = (name, nodeName, pressure) => ({
  name,
  namespace: 'default',
  node: nodeName,
  phase: 'Running',
  ready: true,
  restarts: 0,
  resources: {},
  pressure,
});

const cool = { cpu: 0.1, memory: 0.1, source: 'metrics' };
const hot = { cpu: 0.95, memory: 0.2, source: 'metrics' };

const mockApi = ({ nodes = [], pods = [] }) => {
  global.fetch = jest.fn((url) =>
    Promise.resolve({
      json: () => Promise.resolve(String(url).endsWith('/nodes') ? nodes : pods),
    })
  );
};

afterEach(() => {
  jest.restoreAllMocks();
});

test('renders a sprite per pod and node', async () => {
  mockApi({
    nodes: [node('k-node-1', cool)],
    pods: [pod('api-1', 'k-node-1', cool), pod('api-2', 'k-node-1', cool)],
  });

  render(<App />);

  await waitFor(() => expect(screen.getByAltText('k-node-1')).toBeInTheDocument());
  expect(screen.getByAltText('api-1')).toBeInTheDocument();
  expect(screen.getByAltText('api-2')).toBeInTheDocument();
});

test('shows the fire sprite only on entities over the threshold', async () => {
  mockApi({
    nodes: [node('k-node-1', cool)],
    pods: [pod('cool-pod', 'k-node-1', cool), pod('hot-pod', 'k-node-1', hot)],
  });

  const { container } = render(<App />);

  await waitFor(() => expect(screen.getByAltText('hot-pod')).toBeInTheDocument());
  // Each burning entity renders two layers, so count one of them per entity.
  expect(container.querySelectorAll('img.fire-pod.fire-back')).toHaveLength(1);
  expect(container.querySelectorAll('img.fire-pod.fire-front')).toHaveLength(1);
  expect(container.querySelectorAll('img.fire-node')).toHaveLength(0);
});

// The large flame has to paint behind the entity, and paint order follows
// document order, so its position in the markup is the whole mechanism.
test('draws the back flame before the sprite and the front flame after', async () => {
  mockApi({ nodes: [], pods: [pod('hot-pod', 'k-node-1', hot)] });

  const { container } = render(<App />);

  await waitFor(() => expect(screen.getByAltText('hot-pod')).toBeInTheDocument());
  const order = [...container.querySelectorAll('img')]
    .filter((el) => el.classList.contains('fire-back') || el.classList.contains('fire-front') || el.alt === 'hot-pod')
    .map((el) => (el.alt === 'hot-pod' ? 'sprite' : el.classList.contains('fire-back') ? 'back' : 'front'));

  expect(order).toEqual(['back', 'sprite', 'front']);
});

test('sets a node on fire when it is over the threshold', async () => {
  mockApi({ nodes: [node('k-node-1', hot)], pods: [] });

  const { container } = render(<App />);

  await waitFor(() => expect(screen.getByAltText('k-node-1')).toBeInTheDocument());
  expect(container.querySelectorAll('img.fire-node.fire-back')).toHaveLength(1);
  expect(container.querySelectorAll('img.fire-node.fire-front')).toHaveLength(1);
});

test('survives an API failure without crashing', async () => {
  jest.spyOn(console, 'error').mockImplementation(() => {});
  global.fetch = jest.fn(() => Promise.reject(new Error('connection refused')));

  const { container } = render(<App />);

  await waitFor(() => expect(console.error).toHaveBeenCalled());
  expect(container.querySelectorAll('img.pod')).toHaveLength(0);
});
