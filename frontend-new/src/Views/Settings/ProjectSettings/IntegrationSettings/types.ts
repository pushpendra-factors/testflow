import React from 'react';

export interface IntegrationConfig {
  name: string;
  desc: string;
  icon: string;
  kbLink?: string;
  featureName: string;
  Component: React.ComponentType<any>;
}
