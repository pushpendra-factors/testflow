import { FEATURES, PLANS } from 'Constants/plans.constants';

export interface FeatureConfigState {
  activeFeatures: FeatureConfig[];
  loading: boolean;
  error: boolean;
  addOns?: FeatureConfig[];
  lastRenewedOn?: string;
  plan?: Omit<Plan, 'feature_list'>;
  sixSignalInfo?: SixSignalInfo;
}

export interface FeatureConfig {
  expiry: number;
  granularity: string;
  limit: number;
  is_connected?: boolean;
  is_enabled_feature: boolean;
  name: typeof FEATURES[keyof typeof FEATURES];
}

export enum FeatureConfigActionType {
  UPDATE_FEATURE_CONFIG = 'UPDATE_FEATURE_CONFIG',
  RESET_FEATURE_CONFIG = 'RESET_FEATURE_CONFIG',
  SET_FEATURE_CONFIG_LOADING = 'SET_LOADING',
  SET_FEATURE_CONFIG_ERROR = 'SET_FEATURE_CONFIG_ERROR'
}

interface updateActiveFeatures {
  type: FeatureConfigActionType.UPDATE_FEATURE_CONFIG;
  payload: {
    activeFeatures: FeatureConfig[];
    addOns?: FeatureConfig[];
    lastRenewedOn?: string;
    plan?: Omit<Plan, 'feature_list'>;
    sixSignalInfo?: SixSignalInfo;
  };
}

interface resetActiveFeatures {
  type: FeatureConfigActionType.RESET_FEATURE_CONFIG;
}

interface setFeatureConfigLoading {
  type: FeatureConfigActionType.SET_FEATURE_CONFIG_LOADING;
}

interface setFeatureConfigError {
  type: FeatureConfigActionType.SET_FEATURE_CONFIG_ERROR;
}

export type FeatureConfigActions =
  | updateActiveFeatures
  | resetActiveFeatures
  | setFeatureConfigLoading
  | setFeatureConfigError;

export interface FeatureConfigApiResponse {
  status: number;
  ok: boolean;
  data?: ResponseData;
}

interface ResponseData {
  project_id: number;
  plan: Omit<Plan, 'display_name'>;
  add_ons?: FeatureConfig[];
  last_renewed_on: string;
  six_signal_info: SixSignalInfo;
  display_name?: string;
}

interface SixSignalInfo {
  is_enabled: boolean;
  usage: number;
  limit: number;
}
export interface Plan {
  id: number;
  name: string;
  feature_list?: FeatureConfig[] | null;
  display_name?: string;
}
