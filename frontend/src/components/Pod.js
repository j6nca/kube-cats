import cat from '../assets/pods/cat1.gif'

function Pod({pod}) {
  return <div className="pod">
    <img src={cat}/>
    <span>{pod.metadata.name}</span>
  </div>;
}

export default Pod;
