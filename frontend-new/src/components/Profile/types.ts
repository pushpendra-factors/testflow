export interface DataMap {
  [key: string]: number;
}

export interface ChartProps {
  data: DataMap;
}

export interface TopPage {
  page_url: string;
  views: number;
  users_count: number;
  total_time: number;
  avg_scroll_percent: number;
}

export interface TopUser {
  name: string;
  num_page_views: number;
  active_time: number;
  num_of_pages: number;
}

export type Overview = {
  temperature: number;
  engagement: string;
  users_count: number;
  time_active: number;
  scores_list: DataMap;
  top_pages: TopPage[];
  top_users: TopUser[];
};

export interface AccountOverviewProps {
  overview: Overview;
  loading: boolean;
  top_engagement_signals: string;
}

export interface EventIconProps {
  icon: string;
  size: number;
}

export interface CustomStyles {
  '--icon-size'?: string;
  '--max-lines'?: number;
}

export interface UsernameWithIconProps {
  title: string;
  userID: string;
  isGroupUser: boolean;
  isAnonymous: boolean;
}
export interface TimelineEvent {
  event_name: string;
  alias_name: string;
  display_name: string;
  event_type: string;
  icon: string;
  timestamp: number;
  enabled: boolean;
  is_group_event?: boolean;
  properties?: { [key: string]: unknown };
  username: string;
  user_id: string;
  user_properties?: { [key: string]: unknown };
}

export interface NewEvent {
  id: string;
  name: string;
  display_name: string;
  alias_name?: string;
  icon: string;
  type: string;
  timestamp: number;
  username: string;
  user_id: string;
  is_group_user: boolean;
  is_anonymous_user: boolean;
  properties?: { [key: string]: unknown };
  enabled: boolean;
}

export interface EventDrawerProps {
  visible: boolean;
  event: NewEvent;
  eventPropsType: { [key: string]: string };
  userPropsType: { [key: string]: string };
  onClose: () => void;
}

export interface AccountDrawerProps {
  domain: string;
  visible: boolean;
  onClose: () => void;
  onClickMore: () => void;
  onClickOpenNewtab: () => void;
}

export interface TimelineUser {
  name: string;
  id: string;
  isAnonymous: boolean;
  extraProp?: string;
}

export interface TableRowProps {
  event: NewEvent;
  eventPropsType: { [key: string]: string };
  onEventClick: (event: NewEvent) => void;
}

export interface AccountTimelineTableViewProps {
  timelineEvents?: NewEvent[];
  loading: boolean;
  eventPropsType: { [key: string]: string };
  userPropsType: { [key: string]: string };
  extraClass?: string;
  eventDrawerVisible: boolean;
  setEventDrawerVisible: (value: boolean) => void;
  hasScrollAction: boolean;
  setScrollPercent: (value: number) => void;
  isPreview: boolean;
}
export interface EventDetailsProps {
  event: NewEvent;
  eventPropsType: { [key: string]: string };
  onUpdate: (newOrder: string[]) => void;
}

export interface UserDetailsProps {
  user: TimelineUser;
  userPropsType: { [key: string]: string };
  onUpdate: (newOrder: string[]) => void;
}

export interface TableWithHeadingProps {
  heading: string;
  data: any;
  columns: any;
  xScroll: number;
  yScroll: number;
}
