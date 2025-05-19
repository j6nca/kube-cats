import React, { useState, useEffect } from "react";
import Node from './components/Node'
const localNodes = require('./example/nodes.json')
const localPods = require('./example/pods.json')


const App = () => {
  const [pods, setPods] = useState([]);
  const [nodes, setNodes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const workspace = "dev"
  const endpoints = {
    "prod" : {
      "base" : "http://localhost:8087",
      "nodes_path" : "/api/v1/nodes?limit=500",
      "pods_path" : "/api/v1/pods?limit=500",
    },
    
    "dev" : {
      "base" : "http://localhost:3030",
      "nodes_path" : "/nodes",
      "pods_path" : "/pods",
    },
  }

  // Function to fetch pods and nodes using fetch
  const fetchKubernetesData = async () => {
    try {
      setLoading(true);
      setError(null);

      // Fetch pods data
      const podsResponse = await fetch(
        endpoints[workspace].base + endpoints[workspace].pods_path,
        {}
      );

      // Check if response is okay
      if (!podsResponse.ok) {
        throw new Error("Failed to fetch pods");
      }
      const podsData = await podsResponse.json();

      // Fetch nodes data
      const nodesResponse = await fetch(
        endpoints[workspace].base + endpoints[workspace].nodes_path,
        {}
      );

      // Check if response is okay
      if (!nodesResponse.ok) {
        throw new Error("Failed to fetch nodes");
      }
      const nodesData = await nodesResponse.json();

      setPods(podsData.items);
      setNodes(nodesData.items);
    } catch (error) {
      try{
        if(workspace == "dev"){
          const nodesFile = await fetch('example/nodes.json');
          const podsFile = await fetch('example/pods.json');
          if (!nodesFile.ok || !podsFile.ok) throw new Error('Local file response was not ok');
          const localNodes = await nodesFile.json();
          const localPods = await podsFile.json();
          setPods(localPods.items);
          setNodes(localNodes.items);
        }else{
          throw new Error('Not in dev mode, will not resort to local files');
        }
      }catch (error) {
        setError(error.message);
        console.error("Error fetching data from Kubernetes API:", error);
      }
    } finally {
      setLoading(false);
    }
      
  };

  // Fetch data on component mount
  useEffect(() => {
    fetchKubernetesData();

    // Set up the interval to query the Kubernetes API every 5 minutes
    const intervalId = setInterval(fetchKubernetesData, 300000); // 5 minutes in ms

    // Cleanup on unmount
    return () => clearInterval(intervalId);
  }, []);

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;

  return (
    <div>
      <h1>kube-cats</h1>
        {nodes.map((node) => (
            <Node node={node} pods={pods.filter(pod => pod.spec.nodeName == node.metadata.name)}/>
        ))}
    </div>
  );
};

export default App;
