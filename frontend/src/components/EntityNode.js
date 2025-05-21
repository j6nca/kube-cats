import EntityPod from './EntityPod'
import './Entity.css'
import ufo from '../assets/nodes/ufo.gif'

function EntityNode({node, pods, index}) {
  return <div className={`node fly-animation`} style={{ zIndex: `${index * 2}` }}>
    
    <img title={node.metadata.name} className="node" src={ufo}/>
    {pods.map((pod, podIndex) => (
        <EntityPod key={pod.metadata.name || index} pod={pod} nodeIndex={index} podIndex={podIndex}/>
    ))}
    <span className="label">{node.metadata.name}</span>
    <span className="label">({pods.length})</span>
  </div>;
}

export default EntityNode;