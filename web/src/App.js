import React, { useState, useEffect } from "react";
import EntityNode from './components/EntityNode'
import EntityPod from './components/EntityPod'
import Drifter from './components/Drifter'
import ghLogo from './assets/misc/github.svg'
// Restore alongside the show/hide UI toggle below.
// import eye from './assets/misc/eye.svg'
import config from './config'

// const localNodes = require('./example/nodes.json')
// const localPods = require('./example/pods.json')

const nodes_path = "/nodes"
const pods_path = "/pods"

const fetchKubeInfo = async () => {
  console.log("Fetching kube info from ", config.api_url)
  const [nodesResponse, podsResponse] = await Promise.all([
    fetch(config.api_url + nodes_path),
    fetch(config.api_url + pods_path),
  ]);
  // console.log("nodesResponse: ", nodesResponse)
  // console.log("podsResponse: ", podsResponse)
  const nodesJSON = await nodesResponse.json()
  const podsJSON = await podsResponse.json()
  // console.log("nodesJSON: ", nodesJSON)
  // console.log("podsJSON: ", podsJSON)
  return {
    nodes: nodesJSON,
    pods: podsJSON,
  };
};


const App = () => {
  const [pods, setPods] = useState([]);
  const [nodes, setNodes] = useState([]);
  // Show/hide UI toggle, disabled along with its controls in the markup below.
  // const [showUI, setShowUI] = useState(true);
  // const onClick = () => setShowUI(!showUI)
  useEffect(() => {
    const fetchData = async () => {
      try {
        const { nodes, pods } = await fetchKubeInfo();
        setNodes(nodes);
        setPods(pods);
        console.log("nodes: ", nodes)
        console.log("pods: ", pods)
      } catch (err) {
        console.error("Error fetching Kubernetes data:", err);
      }
    };

    fetchData();
    const interval = setInterval(fetchData, config.refresh_interval_ms);
    return () => clearInterval(interval);
  }, []);

  return (
    <div>
      <div className="watermark">
        {/* {showUI ? <h1>kube-cats 🐱</h1> : null} */}
        {/* <img className="icon show-ui" src={eye} onClick={onClick}></img> */}
        <a href="https://github.com/j6nca/kube-cats">
          <img className="icon" src={ghLogo} alt="kube-cats on GitHub"></img>
        </a>
      </div>
    {pods.map((pod, podIndex) => (
            (podIndex < config.max_pods) ?
            // Enable both directions
            // initial={{ x: (index % 2) * window.innerWidth + (index % 2 > 0 ? 1 : -1) * Math.random() * window.innerWidth * 3, y: Math.random() * window.innerHeight/4 }}
            // animate={{ x: (index + 1) % 2 * window.innerWidth, y: window.innerHeight + Math.random() * window.innerHeight / 2}}
            <Drifter
              key={pod.name || podIndex}
              id={`${pod.namespace}/${pod.name}`}
              laneFraction={0.8}
            >
              <EntityPod pod={pod} podIndex={podIndex}/>
            </Drifter>

             : null
    ))}
    {nodes.map((node, index) => (
      <Drifter
        key={node.name || index}
        id={node.name || `node-${index}`}
        laneFraction={0.6}
        mirror
      >
        <EntityNode node={node} pods={pods.filter(pod => pod.node === node.name)} index={index}/>
      </Drifter>
    ))}
    </div>
  );
};

export default App;
