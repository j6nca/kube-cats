import cat from '../assets/pods/cat.gif'

function EntityPod({pod, podIndex}) {
  // if bundle with node use: zIndex: `${nodeIndex * 2 + (podIndex % 3 > 0 ? 1 : -1)}`
  return <div className="pod" /*style={{ top: `${podIndex/5}%`, left: `${35 + (podIndex % 2 > 0 ? 1 : -1) * Math.random() * (30 + 5) - 5}%` }}*/>
    <img title={pod.name} className="pod" /*style={{ transform: `scaleX(${podIndex % 2 > 0 ? 1 : -1})` }}*/ src={cat}/>
    <span className="label">{pod.name}</span>
  </div>;
}

export default EntityPod;
