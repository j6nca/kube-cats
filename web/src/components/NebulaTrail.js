import React from 'react';
import './NebulaTrail.css';

// This is rendered inside the same translated Drifter as the cat, so the
// soft tail tracks its position without an additional JS animation loop.
const NebulaTrail = () => (
  <div className="nebula-trail" aria-hidden="true" data-testid="nebula-trail">
    <span className="nebula-trail__mist" />
    <span className="nebula-trail__core" />
  </div>
);

export default React.memo(NebulaTrail);
