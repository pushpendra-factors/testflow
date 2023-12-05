interface PLAN_INFO {
  id: string;
  amount: number;
  name: string;
  externalName: string;
}

interface ADDON_INFO {
  id: string;
  amount: number;
}

export interface PlansConfigState {
  plansConfig: {
    plansDetail: PlansDetailStateInterface[];
    addOnsDetail: AddonsStateInterface[];
    loading: boolean;
    error: boolean;
  };
  currentPlanDetail: {
    loading: boolean;
    error: boolean;
    status?: subscriptionStatus;
    renews_on?: string;
    period?: PeriodUnit;
    plan?: PLAN_INFO;
    addons?: ADDON_INFO[];
  };
}

export enum PlansConfigActionType {
  SET_PLANS_DETAIL_LOADING = 'SET_PLANS_DETAIL_LOADING',
  SET_PLANS_DETAIL_ERROR = 'SET_PLANS_DETAIL_ERROR',
  SET_PLANS_CONFIG_DETAILS = 'SET_PLANS_CONFIG_DETAILS',
  SET_CURRENT_PLAN_LOADING = 'SET_CURRENT_PLAN_LOADING',
  SET_CURRENT_PLAN_ERROR = 'SET_CURRENT_PLAN_ERROR',
  SET_CURRENT_PLAN_DETAILS = 'SET_CURRENT_PLAN_DETAILS'
}

interface setPlansConfigLoading {
  type: PlansConfigActionType.SET_PLANS_DETAIL_LOADING;
}

interface setPlansConfigError {
  type: PlansConfigActionType.SET_PLANS_DETAIL_ERROR;
}

interface setCurrentPlanDetailLoading {
  type: PlansConfigActionType.SET_CURRENT_PLAN_LOADING;
}

interface setCurrentPlanDetailError {
  type: PlansConfigActionType.SET_CURRENT_PLAN_ERROR;
}

interface setPlansConfigDetails {
  type: PlansConfigActionType.SET_PLANS_CONFIG_DETAILS;
  payload: {
    plansDetail: PlansDetailStateInterface[];
    addOnsDetail: AddonsStateInterface[];
  };
}

interface setCurrentPlanDetails {
  type: PlansConfigActionType.SET_CURRENT_PLAN_DETAILS;
  payload: {
    status?: subscriptionStatus;
    renews_on?: string;
    period: number;
    plan?: PLAN_INFO;
    addons?: ADDON_INFO[];
  };
}

export type PlansConfigActions =
  | setPlansConfigDetails
  | setPlansConfigLoading
  | setPlansConfigError
  | setCurrentPlanDetailError
  | setCurrentPlanDetailLoading
  | setCurrentPlanDetails;

interface ApiResponse {
  status: number;
  ok: boolean;
  data?: any;
}

export interface PlansDetailAPIResponse extends ApiResponse {
  data?: PlansDetailsResponse[];
}

export interface SubscriptionDetailsAPIResponse extends ApiResponse {
  data?: SubscriptionDeatilsResponse;
}

interface PlansDetailsResponse {
  type: planTypes;
  name: string;
  external_name: string;
  id: string;
  price: number;
  period_unit: PeriodUnit;
}

type planTypes = 'plan' | 'addon';
type PeriodUnit = 'month' | 'year';

type subscriptionStatus =
  | 'future'
  | 'in_trial'
  | 'active'
  | 'non_renewing'
  | 'paused'
  | 'cancelled';

interface SubscriptionDeatilsResponse {
  status: subscriptionStatus;
  renews_on: string;
  period_unit: PeriodUnit;
  subscription_details: {
    type: planTypes;
    id: string;
    amount: number;
    external_name: string;
  }[];
}

export interface PlansDetailStateInterface {
  name: string;
  terms: PlanTerm[];
}

export interface PlanTerm {
  id: string;
  price: number;
  period: PeriodUnit;
  name: string;
}

export interface AddonsStateInterface {
  name: string;
  id: string;
  price: number;
}

export interface GetInvoicesAPIResponse extends ApiResponse {
  data?: Invoice[];
}

export interface Invoice {
  id: string;
  billing_date: string;
  Amount: number;
  amount_paid: number;
  AmountDue: number;
  items: string[];
}
