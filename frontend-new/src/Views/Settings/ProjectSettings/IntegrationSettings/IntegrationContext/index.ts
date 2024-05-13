import React from 'react';
import { IntegrationContextData } from '../types';

export const defaultIntegrationContextData: IntegrationContextData = {
  integrationStatus: {},
  dataLoading: false,
  integrationStatusLoading: false
};
export const IntegrationContext = React.createContext(
  defaultIntegrationContextData
);
