import cat from '../assets/pods/cat1.gif'

function EntityPod({pod, nodeIndex, podIndex}) {
  return <div className="pod" style={{ top: `${podIndex/5}%`, left: `${35 + (podIndex % 2 > 0 ? 1 : -1) * Math.random() * (30 + 5) - 5}%`, zIndex: `${nodeIndex * 2 + (podIndex % 3 > 0 ? 1 : -1)}` }}>
    <img title={pod.metadata.name} className="pod" style={{ transform: `scaleX(${podIndex % 2 > 0 ? 1 : -1})` }} src={cat}/>
    <span className="label">{pod.metadata.name}</span>
  </div>;
}

export default EntityPod;
