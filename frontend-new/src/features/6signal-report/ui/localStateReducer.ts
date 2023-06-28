import logger from 'Utils/logger';
import { WeekStartEnd, ShareData, ReportApiResponseData } from '../types';

// Action Types
export enum VisitorReportActions {
  SET_DRAWER_VISIBILITY = 'SET_DRAWER_VISIBILITY',
  REPORT_DATA_LOADING = 'REPORT_DATA_LOADING',
  REPORT_DATA_ERROR = 'REPORT_DATA_ERROR',
  REPORT_DATA_LOADED = 'REPORT_DATA_LOADED',
  SET_PARSED_VALUES = 'SET_PARSED_VALUES',
  SET_CAMPAIGNS = 'SET_CAMPAIGNS',
  SET_SELECTED_CAMPAIGNS = 'SET_SELECTED_CAMPAIGNS',
  SET_CHANNELS = 'SET_CHANNELS',
  SET_SELECTED_CHANNELS = 'SET_SELECTED_CHANNELS',
  SET_DATE_VALUES = 'SET_DATE_VALUES',
  SET_SELECTED_DATE = 'SET_SELECTED_DATE',
  SET_CAMPAIGN_SELECT_VISIBILITY = 'SET_CAMPAIGN_SELECT_VISIBILITY',
  SET_CHANNEL_SELECTION_VISIBILITY = 'SET_CHANNEL_SELECTION_VISIBILITY',
  SET_DATE_SELECTION_VISIBILITY = 'SET_DATE_SELECTION_VISIBILITY',
  SET_SHARE_DATA_LOADING = 'SET_SHARE_DATA_LOADING',
  SET_SHARE_DATA_FAILED = 'SET_SHARE_DATA_FAILED',
  SET_SHARE_DATA = 'SET_SHARE_DATA',
  SET_PAGE_MODE = 'SET_PAGE_MODE',
  RESET_FILTERS = 'RESET_FILTERS',
  SET_SHARE_MODAL_VISIBILITY = 'SET_SHARE_MODAL_VISIBILITY',
  SET_PAGE_URL_DATA = 'SET_PAGE_URL_DATA',
  SET_PAGE_URL_DATA_LOADING = 'SET_PAGE_URL_DATA_LOADING',
  SET_PAGE_URL_DATA_ERROR = 'SET_PAGE_URL_DATA_ERROR',
  SET_PAGE_VIEW_SELECTION_VISIBILITY = 'SET_PAGE_VIEW_SELECTION_VISIBILITY',
  SET_SELECTED_PAGE_VIEWS = 'SET_SELECTED_PAGE_VIEWS',
  SET_PAST_DATE_DATA_AVAILABILITY = 'SET_PAST_DATE_DATA_AVAILABILITY'
}

//Action Types and payload

interface SetBooleanPayload {
  type:
    | VisitorReportActions.SET_DRAWER_VISIBILITY
    | VisitorReportActions.SET_CAMPAIGN_SELECT_VISIBILITY
    | VisitorReportActions.SET_CHANNEL_SELECTION_VISIBILITY
    | VisitorReportActions.SET_DATE_SELECTION_VISIBILITY
    | VisitorReportActions.SET_SHARE_MODAL_VISIBILITY
    | VisitorReportActions.SET_PAGE_VIEW_SELECTION_VISIBILITY
    | VisitorReportActions.SET_PAST_DATE_DATA_AVAILABILITY;
  payload: boolean;
}

interface SetStateWithoutPayload {
  type:
    | VisitorReportActions.REPORT_DATA_ERROR
    | VisitorReportActions.REPORT_DATA_LOADING
    | VisitorReportActions.RESET_FILTERS
    | VisitorReportActions.SET_SHARE_DATA_LOADING
    | VisitorReportActions.SET_SHARE_DATA_FAILED
    | VisitorReportActions.SET_PAGE_URL_DATA_LOADING
    | VisitorReportActions.SET_PAGE_URL_DATA_ERROR;
}

interface SetReportData {
  type: VisitorReportActions.REPORT_DATA_LOADED;
  payload: ReportApiResponseData;
}

interface SetStringArrayPayload {
  type:
    | VisitorReportActions.SET_CAMPAIGNS
    | VisitorReportActions.SET_SELECTED_CAMPAIGNS
    | VisitorReportActions.SET_CHANNELS
    | VisitorReportActions.SET_PAGE_URL_DATA
    | VisitorReportActions.SET_SELECTED_PAGE_VIEWS;
  payload: string[];
}

interface SetStringPayload {
  type:
    | VisitorReportActions.SET_SELECTED_CHANNELS
    | VisitorReportActions.SET_SELECTED_DATE;

  payload: string;
}

interface SetDateValues {
  type: VisitorReportActions.SET_DATE_VALUES;
  payload: WeekStartEnd[];
}

interface SetShareData {
  type: VisitorReportActions.SET_SHARE_DATA;
  payload: ShareData;
}

interface SetPageMode {
  type: VisitorReportActions.SET_PAGE_MODE;
  payload: PageMode;
}

interface SetParsedValues {
  type: VisitorReportActions.SET_PARSED_VALUES;
  payload: {
    campaigns: string[];
    channels: string[];
  };
}

//State Type
interface ReportDataType {
  data: ReportApiResponseData | null;
  loading: boolean;
  error: boolean;
  isNotInitialized?: boolean;
}

type PageMode = 'in-app' | 'public';

interface ShareDataType {
  data: ShareData | null;
  loading: boolean;
  error: boolean;
}

interface PageViewUrls {
  data: string[] | null;
  loading: boolean;
  error: boolean;
}

interface VisitorReportState {
  drawerVisible: boolean;
  reportData: ReportDataType;
  campaigns: string[];
  selectedCampaigns: string[];
  channels: string[];
  selectedChannel: string;
  dateValues: WeekStartEnd[];
  selectedDate: string;
  selectedPageViews: string[];
  campaignSelectionVisibility: boolean;
  channelSelectionVisibility: boolean;
  dateSelectionVisibility: boolean;
  pageViewSelectionVisibility: boolean;
  pageMode: PageMode;
  shareData: ShareDataType;
  shareModalVisibility: boolean;
  pageViewUrls: PageViewUrls;
  isPastDatesDataAvailable: boolean;
}

type Action =
  | SetBooleanPayload
  | SetStateWithoutPayload
  | SetStringPayload
  | SetStringArrayPayload
  | SetReportData
  | SetDateValues
  | SetShareData
  | SetPageMode
  | SetParsedValues;

export const initialState: VisitorReportState = {
  drawerVisible: false,
  reportData: {
    data: null,
    error: false,
    loading: false,
    isNotInitialized: true
  },
  campaigns: [],
  selectedCampaigns: [],
  channels: [],
  selectedChannel: '',
  dateValues: [],
  selectedDate: '',
  campaignSelectionVisibility: false,
  channelSelectionVisibility: false,
  dateSelectionVisibility: false,
  pageViewSelectionVisibility: false,
  selectedPageViews: [],
  shareModalVisibility: false,
  pageMode: 'in-app',
  shareData: {
    data: null,
    loading: false,
    error: false
  },
  pageViewUrls: {
    data: null,
    loading: false,
    error: false
  },
  isPastDatesDataAvailable: false
};

export function visitorReportReducer(
  state: VisitorReportState,
  action: Action
): VisitorReportState {
  switch (action.type) {
    case VisitorReportActions.SET_DRAWER_VISIBILITY:
      return {
        ...state,
        drawerVisible: action.payload
      };
    case VisitorReportActions.REPORT_DATA_ERROR:
      return {
        ...state,
        reportData: {
          error: true,
          data: null,
          loading: false
        }
      };
    case VisitorReportActions.REPORT_DATA_LOADING:
      return {
        ...state,
        reportData: {
          ...state.reportData,
          loading: true
        }
      };
    case VisitorReportActions.REPORT_DATA_LOADED:
      return {
        ...state,
        reportData: {
          loading: false,
          error: false,
          data: action.payload
        }
      };
    case VisitorReportActions.SET_CAMPAIGNS:
      return {
        ...state,
        campaigns: action.payload
      };
    case VisitorReportActions.SET_SELECTED_CAMPAIGNS:
      return {
        ...state,
        selectedCampaigns: action.payload,
        campaignSelectionVisibility: false
      };
    case VisitorReportActions.SET_CHANNELS:
      return {
        ...state,
        channels: action.payload
      };
    case VisitorReportActions.SET_SELECTED_CHANNELS:
      return {
        ...state,
        selectedChannel: action.payload
      };
    case VisitorReportActions.SET_DATE_VALUES:
      return {
        ...state,
        dateValues: action.payload
      };
    case VisitorReportActions.SET_SELECTED_DATE:
      return {
        ...state,
        selectedDate: action.payload,
        dateSelectionVisibility: false
      };
    case VisitorReportActions.SET_CAMPAIGN_SELECT_VISIBILITY:
      return {
        ...state,
        campaignSelectionVisibility: action.payload
      };
    case VisitorReportActions.SET_CHANNEL_SELECTION_VISIBILITY:
      return {
        ...state,
        channelSelectionVisibility: action.payload
      };
    case VisitorReportActions.SET_DATE_SELECTION_VISIBILITY: {
      return {
        ...state,
        dateSelectionVisibility: action.payload
      };
    }
    case VisitorReportActions.SET_SHARE_DATA:
      return {
        ...state,
        shareData: {
          loading: false,
          error: false,
          data: action.payload
        }
      };
    case VisitorReportActions.SET_SHARE_DATA_FAILED:
      return {
        ...state,
        shareData: {
          data: null,
          error: true,
          loading: false
        }
      };
    case VisitorReportActions.SET_SHARE_DATA_LOADING:
      return {
        ...state,
        shareData: {
          ...state.shareData,
          loading: true
        }
      };
    case VisitorReportActions.SET_PAGE_MODE:
      return {
        ...state,
        pageMode: action.payload
      };
    case VisitorReportActions.RESET_FILTERS:
      return {
        ...state,
        campaigns: [],
        channels: [],
        selectedCampaigns: [],
        selectedChannel: '',
        selectedPageViews: []
      };
    case VisitorReportActions.SET_SHARE_MODAL_VISIBILITY:
      return {
        ...state,
        shareModalVisibility: action.payload
      };
    case VisitorReportActions.SET_PARSED_VALUES:
      return {
        ...state,
        campaigns: action.payload.campaigns,
        channels: action.payload.channels
      };
    case VisitorReportActions.SET_PAGE_URL_DATA:
      return {
        ...state,
        pageViewUrls: {
          data: action.payload,
          loading: false,
          error: false
        }
      };
    case VisitorReportActions.SET_PAGE_URL_DATA_ERROR:
      return {
        ...state,
        pageViewUrls: {
          data: null,
          loading: false,
          error: true
        }
      };
    case VisitorReportActions.SET_PAGE_URL_DATA_LOADING:
      return {
        ...state,
        pageViewUrls: {
          ...state.pageViewUrls,
          loading: true
        }
      };
    case VisitorReportActions.SET_PAGE_VIEW_SELECTION_VISIBILITY:
      return {
        ...state,
        pageViewSelectionVisibility: action.payload
      };
    case VisitorReportActions.SET_SELECTED_PAGE_VIEWS:
      return {
        ...state,
        selectedPageViews: action.payload,
        pageViewSelectionVisibility: false
      };
    case VisitorReportActions.SET_PAST_DATE_DATA_AVAILABILITY:
      return {
        ...state,
        isPastDatesDataAvailable: action.payload
      };
    default:
      logger.error('Unsupported visitor action type');
      return { ...state };
  }
}
