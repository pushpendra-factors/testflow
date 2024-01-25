import { QUERY_TYPE_EVENT } from "Utils/constants";
import { BaseStateClass } from "Types/BaseStateClass";

export interface QueryParams {
    query_id: string,
    query_type: string
}

export interface GroupByType {
    prop_category: string, // user / event
    property: string, // user/eventproperty
    prop_type: string, // categorical  /numberical
    eventValue: string, // event name (funnel only)
    eventName: string, // eventName $present for global user breakdown
    eventIndex: number
  }

export interface GroupBy {
    global?: Array<GroupByType>
    event?: Array<GroupByType>
}

export class ResultState {
    loading: boolean = false;
    error: boolean = false;
    data: any = null;
    apiCallStatus: any = { required: true, message: null }
  };

export class QueryOptions {
    group_analysis: string = 'users'
    groupBy: GroupBy | any[] = {}
    globalFilters: Array<any> = []
    event_analysis_seq: string = ''
    session_analytics_seq: {} = {}
    date_range: {} = {}
    events_condition: string = 'any_given_event'
  };

export class CoreQueryState extends BaseStateClass {
    queryType: string = QUERY_TYPE_EVENT;
    querySaved: any = false;
    requestQuery: any = null;
    loading: boolean = true;
    showResult: boolean = false;
    queries: Array<any> = [];
    appliedQueries: Array<any> = [];
    queryOptions: QueryOptions = new QueryOptions();
    appliedBreakdown: Array<any> = [];
    resultState: ResultState = new ResultState();
    activeTab: number = 1;

    constructor(){super()}
}
