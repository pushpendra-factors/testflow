export interface EventIconProps {
  icon: string;
  size: number;
}

export interface CustomStyles {
  '--bg-color': string;
  '--border-color': string;
  '--icon-size': string;
}

export interface UsernameWithIconProps {
  title: string;
  userID: string;
  isAnonymous: boolean;
}

export interface EventDrawerProps {
  visible: boolean;
  event: TimelineEvent;
  onClose: () => void;
}

interface TimelineEvent {
  timestamp: number;
  icon: string;
  event_name: string;
  alias_name: string;
  display_name: string;
  event_type: string;
  properties?: { [key: string]: any };
  user: string;
  id: string;
}

export interface TableRowProps {
  event: TimelineEvent;
  user: TimelineUser;
  onEventClick: (event: TimelineEvent) => void;
}

interface TimelineUser {
  title: string;
  subtitle: string;
  userId: string;
  isAnonymous: boolean;
}
export interface AccountTimelineTableViewProps {
  timelineEvents?: TimelineEvent[];
  timelineUsers?: TimelineUser[];
  loading: boolean;
}
