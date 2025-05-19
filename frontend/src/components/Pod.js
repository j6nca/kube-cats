function Pod({pod}) {
  return <div>
      <li key={pod.metadata.name}>
      {pod.metadata.name}
        <ul>
          <li>
            {pod.status.phase}
          </li>
          <li>
            {pod.metadata.namespace}
          </li>
        </ul>
      </li>
  </div>;
}

export default Pod;
