// import EntityPod from './EntityPod'
import './Entity.css'
import ufo from '../assets/nodes/ufo.gif'
import fire from '../assets/misc/fire.svg'
import config from '../config'
import { isHot, nodeTooltip } from '../heat'

function EntityNode({node, pods, index}) {
  const hot = isHot(node.pressure)
  const title = config.show_tooltips ? nodeTooltip(node, pods.length) : node.name

  return <div className={`node`} style={{ zIndex: 1 }}>

    <img title={title} className="node" src={ufo} alt={node.name}/>

    {hot ? <img className="fire fire-node" src={fire} alt="" aria-hidden="true"/> : null}

    {/* Enable to draw pods with nodes */
    /* {pods.map((pod, podIndex) => (
            (podIndex < config.max_pods) ?
            <EntityPod key={pod.name || index} pod={pod} nodeIndex={index} podIndex={podIndex}/>
             : null
    ))} */}
    {config.show_labels ? <span className="label">{node.name}</span> : null}
    {config.show_pod_count ? <span className="label"> ({pods.length})</span> : null}
  </div>;
}

export default EntityNode;
