export interface EventIconProps {
  icon: string;
  size: number;
}

export interface CustomStyles {
  '--bg-color': string;
  '--border-color': string;
}

export interface UsernameWithIconProps {
  title: string;
  userID: string;
  isAnonymous: boolean;
}

export interface EventDrawerProps {
  visible: boolean;
  selectedEvent: any;
  onClose: () => void;
}

interface Event {
  timestamp: number;
  icon: string;
  event_name: string;
  alias_name: string;
  display_name: string;
  properties?: { [key: string]: string };
  user: string;
  id: string;
}

interface User {
  title: string;
  subtitle: string;
  userId: string;
  isAnonymous: boolean;
}

export interface TableRowProps {
  event: Event;
  user: User;
  onEventClick: (event: Event) => void;
}

interface TimelineEvent {
  timestamp: number;
  icon: string;
  eventName: string;
  properties?: { [key: string]: string };
  user: string;
  id: string;
}

interface TimelineUser {
  title: string;
  subtitle: string;
  userId: string;
  isAnonymous: boolean;
}

interface EventNamesMap {
  [key: string]: {
    [nestedKey: string]: string;
  };
}

export interface AccountTimelineTableViewProps {
  timelineEvents?: TimelineEvent[];
  timelineUsers?: TimelineUser[];
  loading: boolean;
  eventNamesMap: EventNamesMap;
}
