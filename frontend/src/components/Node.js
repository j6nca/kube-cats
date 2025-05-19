import Pod from './Pod'
function Node({node, pods}) {
  return <div>
    <ul>
        <li key={node.metadata.name}>
            {node.metadata.name}
        </li>
        <ul>
            {pods.map((pod) => (
                <li>
                    <Pod pod={pod}/>
                </li>
            ))}
        </ul>
    </ul>
  </div>;
}

export default Node;
