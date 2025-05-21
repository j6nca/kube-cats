import Pod from './Pod'
import './kube-cats.css'
import React, { useState, useEffect } from 'react';
import ufo from '../assets/nodes/ufo.gif'

function Node({node, pods}) {
  const [isVisible, setIsVisible] = useState(false);
  const [position, setPosition] = useState({ x: 0, y: 0 });

  useEffect(() => {
    setIsVisible(true);

    // Set a timeout to trigger the animation after a short delay
    const timeoutId = setTimeout(() => {
      setPosition({ x: window.innerWidth - 100, y: window.innerHeight - 100 });
    }, 100);

    // Clean up the timeout if the component unmounts
    return () => clearTimeout(timeoutId);
  }, []);
  return <div className={`node fly-animation`}>
    {node.metadata.name}
    <img className="node" src={ufo}/>
    {pods.map((pod) => (
            <Pod pod={pod}/>
    ))}
  </div>;
}

export default Node;
