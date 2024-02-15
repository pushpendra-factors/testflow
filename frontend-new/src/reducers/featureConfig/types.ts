import { FEATURES } from 'Constants/plans.constants';

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
  name: (typeof FEATURES)[keyof typeof FEATURES];
}

export enum FeatureConfigActionType {
  UPDATE_FEATURE_CONFIG = 'UPDATE_FEATURE_CONFIG',
  RESET_FEATURE_CONFIG = 'RESET_FEATURE_CONFIG',
  SET_FEATURE_CONFIG_LOADING = 'SET_LOADING',
  SET_FEATURE_CONFIG_ERROR = 'SET_FEATURE_CONFIG_ERROR',
  SET_PLANS_DETAIL_LOADING = 'SET_PLANS_DETAIL_LOADING',
  SET_ADDONS_DETAIL = 'SET_ADDONS_DETAIL',
  SET_PLANS_DETAIL = 'SET_PLANS_DETAIL'
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

interface ApiResponse {
  status: number;
  ok: boolean;
  data?: any;
}
export interface FeatureConfigApiResponse extends ApiResponse {
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

export interface SixSignalInfo {
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
