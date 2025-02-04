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
    | 'connected'
    | 'success'
    | 'disconnected'
    | 'synced'
    | 'delayed'
    | 'pending'
    | 'large_data_delayed'
    | 'client_side_token_expired'
    | 'limit_exceed';
  last_synced_at: number;
  message?: string;
}

export type IntegrationState =
  | 'connected'
  | 'error'
  | 'pending'
  | 'not_connected';
