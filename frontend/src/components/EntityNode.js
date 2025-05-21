import EntityPod from './EntityPod'
import './Entity.css'
import ufo from '../assets/nodes/ufo.gif'

function EntityNode({node, pods, index}) {
  return <div className={`node fly-animation`} style={{ zIndex: `${index * 2}` }}>
    
    <img className="node" src={ufo}/>
    {pods.map((pod, index) => (
        <EntityPod key={pod.metadata.name || index} pod={pod} index={index}/>
    ))}
    <span className="label">{node.metadata.name}</span>
  </div>;
}

export default EntityNode;