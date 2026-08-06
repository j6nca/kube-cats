import cat_1 from '../assets/pods/cat_1.gif'
import cat_2 from '../assets/pods/cat_2.gif'
import cat_3 from '../assets/pods/cat_3.gif'
import cat_4 from '../assets/pods/cat_4.gif'
import cat_5 from '../assets/pods/cat_5.gif'
import cat_6 from '../assets/pods/cat_6.gif'
import cat_7 from '../assets/pods/cat_7.gif'
import cat_8 from '../assets/pods/cat_8.gif'
import cat_9 from '../assets/pods/cat_9.gif'
import fire from '../assets/misc/fire.svg'
import config from '../config'
import { isHot, podTooltip } from '../heat'

const cats = [
  cat_1,
  cat_2,
  cat_3,
  cat_4,
  cat_5,
  cat_6,
  cat_7,
  cat_8,
  cat_9,
]

function EntityPod({pod, podIndex}) {
  const hot = isHot(pod.pressure)
  const title = config.show_tooltips ? podTooltip(pod) : pod.name

  return <div className="pod" style={{ zIndex: 0 }}>
    <img title={title} className="pod" src={cats[podIndex % cats.length]} alt={pod.name}/>
    {hot ? <img className="fire fire-pod" src={fire} alt="" aria-hidden="true"/> : null}
    {config.show_labels ? <span className="label">{pod.name}</span> : null}
  </div>;
}

export default EntityPod;
