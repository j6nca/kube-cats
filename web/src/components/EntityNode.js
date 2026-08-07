// import EntityPod from './EntityPod'
import './Entity.css'
import ufo from '../assets/nodes/ufo.gif'
import fireBack from '../assets/misc/fire-back.svg'
import fireFront from '../assets/misc/fire-front.svg'
import config from '../config'
import { isHot, nodeTooltip } from '../heat'

function EntityNode({node, pods, index}) {
  const hot = isHot(node.pressure)
  const title = config.show_tooltips ? nodeTooltip(node, pods.length) : node.name

  // The large flame is drawn before the sprite so it rises from behind it, the
  // two small ones after so they lick over the front.
  return <div className={`node`} style={{ zIndex: 1 }}>

    {hot ? <img className="fire fire-node fire-back" src={fireBack} alt="" aria-hidden="true"/> : null}

    <img title={title} className="node" src={ufo} alt={node.name}/>

    {hot ? <img className="fire fire-node fire-front" src={fireFront} alt="" aria-hidden="true"/> : null}

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
