import React from 'react';
import { FEATURES } from 'Constants/plans.constants';

export interface IntegrationConfig {
  id: string;
  categoryId: string;
  name: string;
  desc: string;
  icon: string;
  kbLink?: string;
  featureName: string;
  Component: React.ComponentType<any>;
  instructionTitle?: string;
  instructionDescription?: string;
  showInstructionMenu: boolean;
}

export interface IntegrationCategroryType {
  name: string;
  id: string;
}

export interface IntegrationContextData {
  integrationStatus: IntegrationStatusData;
  dataLoading: boolean;
  integrationStatusLoading: boolean;
}

export interface IntegrationStatusData {
  [key: (typeof FEATURES)[keyof typeof FEATURES]]: IntegrationStatus;
}

export interface IntegrationStatus {
  state:
    | ''
    | 'synced'
    | 'delayed'
    | 'pull_delayed'
    | 'sync_pending'
    | 'heavy_delayed'
    | 'client_token_expired'
    | 'limit_exceed';
  last_synced_at: number;
  message?: string;
}
