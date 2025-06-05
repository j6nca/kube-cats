// import EntityPod from './EntityPod'
import './Entity.css'
import ufo from '../assets/nodes/ufo.gif'
import config from '../config'

function EntityNode({node, pods, index}) {
  // console.log(config)
  return <div className={`node`} style={{ zIndex: 1 }}>
    
    <img title={node.name} className="node" src={ufo}/>
    {/* {pods.map((pod, podIndex) => (
            (podIndex < config.max_pods_per_node) ? 
            <EntityPod key={pod.name || index} pod={pod} nodeIndex={index} podIndex={podIndex} test={config.max_pods_per_node}/>
             : null
    ))} */}
    <span className="label">{node.name}</span>
    <span className="label"> ({pods.length})</span>
  </div>;
}

export default EntityNode;
