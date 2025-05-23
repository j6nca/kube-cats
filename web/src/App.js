import React, { useState, useEffect } from "react";
import EntityNode from './components/EntityNode'
import { motion } from "framer-motion";
import ghLogo from './assets/misc/github.svg'
import eye from './assets/misc/eye.svg'
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
  const [showUI, setShowUI] = useState(true);
  const onClick = () => setShowUI(!showUI)
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
    const interval = setInterval(fetchData, 5 * 60 * 1000); // every 5 minutes
    return () => clearInterval(interval);
  }, []);

  return (
    <div>
      <div className="watermark">
        {showUI ? <h1>kube-cats 🐱</h1> : null}
        <img className="icon show-ui" src={eye} onClick={onClick}></img>
        {showUI ? <a href="https://github.com/j6nca/kube-cats">
          <img className="icon" src={ghLogo}></img>
        </a> : null}
      </div>
    {nodes.map((node, index) => (
      <motion.div
        key={node.metadata.name || index}
        initial={{ x: (index % 2) * window.innerWidth + (index % 2 > 0 ? 1 : -1) * 100, y: Math.random() * window.innerHeight/4 }}
        animate={{ x: (index + 1) % 2 * window.innerWidth, y: window.innerHeight + 100}}
        transition={{ duration: 15, ease: "linear", repeat: Infinity }}
        style={{ position: 'absolute' }}
      >
        <EntityNode node={node} pods={pods.filter(pod => pod.spec.nodeName == node.metadata.name)} index={index}/>
      </motion.div>
    ))}
    </div>
  );
};

export default App;
