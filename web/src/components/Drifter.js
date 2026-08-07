import { motion, useTime, useTransform } from "framer-motion";

export const OFFSCREEN_OFFSET = 500;

// FNV-1a, mapped to [0, 1).
//
// Entity parameters are derived from the entity's name rather than
// Math.random() so a pod keeps the same lane, speed and phase for as long as it
// exists. The previous inline Math.random() calls re-rolled on every React
// render, so every refetch teleported the whole cluster.
export const hashUnit = (str) => {
  let hash = 2166136261;
  for (let i = 0; i < str.length; i += 1) {
    hash ^= str.charCodeAt(i);
    hash = Math.imul(hash, 16777619);
  }
  return (hash >>> 0) / 4294967296;
};

// Horizontal position at a point in time, as a pure function so the wrapping
// behaviour can be tested without a renderer.
//
// `phase` is the entity's starting offset along its own path, in [0, 1). It is
// the whole trick behind not looking like a swarm: without it every entity is
// at progress 0 on the first frame and they all launch from the left edge
// together.
export const driftX = ({ elapsedMs, duration, phase, width, mirror = false }) => {
  const span = width + OFFSCREEN_OFFSET * 2;
  const elapsedCycles = elapsedMs / 1000 / duration + phase;

  let progress;
  if (mirror) {
    // Out and back, one leg per `duration`, matching the old repeatType.
    const cycle = ((elapsedCycles % 2) + 2) % 2;
    progress = cycle <= 1 ? cycle : 2 - cycle;
  } else {
    progress = ((elapsedCycles % 1) + 1) % 1;
  }

  return -OFFSCREEN_OFFSET + progress * span;
};

// Per-entity motion, stable for a given id.
export const driftParams = (id) => ({
  duration: 10 + hashUnit(`${id}:duration`) * 30,
  phase: hashUnit(`${id}:phase`),
  lane: hashUnit(`${id}:lane`),
});

// Drifts a sprite across the viewport, wrapping forever.
//
// Driven from a time value rather than a repeating keyframe animation: a
// repeating animation always replays from its `initial` x, so seeding it with a
// random start would make the sprite teleport back to mid-screen on every loop.
const Drifter = ({ id, laneFraction, mirror = false, children }) => {
  const time = useTime();
  const { duration, phase, lane } = driftParams(id);

  const x = useTransform(time, (elapsedMs) =>
    driftX({ elapsedMs, duration, phase, width: window.innerWidth, mirror })
  );

  return (
    <motion.div style={{ position: "absolute", x, y: lane * window.innerHeight * laneFraction }}>
      {children}
    </motion.div>
  );
};

export default Drifter;
