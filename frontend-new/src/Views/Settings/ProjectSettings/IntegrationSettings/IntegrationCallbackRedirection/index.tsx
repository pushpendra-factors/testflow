import { PathUrls } from 'Routes/pathUrls';
import React from 'react';
import { Redirect, useLocation, useParams } from 'react-router-dom';

const IntegrationRedirection = () => {
  const { integration_id: integrationId } = useParams();
  const location = useLocation();

  return (
    <Redirect
      to={`${PathUrls.SettingsIntegration}/${integrationId}${location?.search}`}
    />
  );
};

export default IntegrationRedirection;
