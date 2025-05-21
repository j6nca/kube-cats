import cat from '../assets/pods/cat1.gif'

function EntityService({pod, index}) {
  return <div className="pod" style={{ top: `${Math.random() * 5}%`, left: `${Math.random() * 50}%`, zIndex: `${index * 2 + 1}` }}>
    <img className="pod" src={cat}/>
    <span className="label">{pod.metadata.name}</span>
  </div>;
}

export default EntityService;
