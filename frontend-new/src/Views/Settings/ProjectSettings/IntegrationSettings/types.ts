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
  sortOrder: number;
}

export interface IntegrationContextData {
  integrationStatus: IntegrationStatusData;
  dataLoading: boolean;
  integrationStatusLoading: boolean;
  fetchIntegrationStatus?: () => void;
}

export interface IntegrationStatusData {
  [key: (typeof FEATURES)[keyof typeof FEATURES]]: IntegrationStatus;
}

export interface IntegrationStatus {
  state:
    | ''
    | 'connected'
    | 'success'
    | 'disconnected'
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
