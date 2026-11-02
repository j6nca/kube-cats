import React from 'react';
import './NebulaTrail.css';

// A single ambient stream through the shared flight corridor, not a per-cat
// exhaust trail. All motion is CSS-only and independent of pod updates.
const NebulaTrail = () => (
  <div className="nebula-stream" aria-hidden="true" data-testid="nebula-trail">
    <div className="nebula-stream__haze" />
    <div className="nebula-stream__ribbon" />
  </div>
);

export default React.memo(NebulaTrail);
