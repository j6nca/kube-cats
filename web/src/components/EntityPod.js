import cat_1 from '../assets/pods/cat_1.gif'
import cat_2 from '../assets/pods/cat_2.gif'
import cat_3 from '../assets/pods/cat_3.gif'
import cat_4 from '../assets/pods/cat_4.gif'
import cat_5 from '../assets/pods/cat_5.gif'
import cat_6 from '../assets/pods/cat_6.gif'
import cat_7 from '../assets/pods/cat_7.gif'
import cat_8 from '../assets/pods/cat_8.gif'
import cat_9 from '../assets/pods/cat_9.gif'

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
  return <div className="pod" style={{ zIndex: 0/*zIndex: `${nodeIndex * 2 + (podIndex % 3 > 0 ? 1 : -1)}` , top: `${podIndex/5}%`, left: `${35 + (podIndex % 2 > 0 ? 1 : -1) * Math.random() * (30 + 5) - 5}%`*/}}>
    <img title={pod.name} className="pod" /*style={{ transform: `scaleX(${podIndex % 2 > 0 ? 1 : -1})` }}*/ src={cats[podIndex % cats.length]}/>
    <span className="label">{pod.name}</span>
  </div>;
}

export default EntityPod;
