import React, { useEffect } from 'react';
import AttributionSetupDone from './AttributionSetupDone';
import AttributionSetupPending from './AttributionSetupPending';

function AttributionBaseComponent() {
  useEffect(() => {
    // implement logic for redirections
  }, []);

  const setupDone = true;
  if (setupDone) return <AttributionSetupDone />;
  return <AttributionSetupPending />;
}

export default AttributionBaseComponent;
