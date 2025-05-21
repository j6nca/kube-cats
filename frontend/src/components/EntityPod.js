import cat from '../assets/pods/cat1.gif'

function EntityPod({pod, nodeIndex, podIndex}) {
  return <div className="pod" style={{ top: `${Math.random() * 5}%`, left: `${Math.random() * 50}%`, zIndex: `${nodeIndex * 2 + (podIndex % 2 > 0 ? 1 : -1)}` }}>
    <img title={pod.metadata.name} className="pod" src={cat}/>
    <span className="label">{pod.metadata.name}</span>
  </div>;
}

export default EntityPod;
