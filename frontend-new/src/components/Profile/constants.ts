import { TimelineConfig } from './types';

export const GROUP_TO_DOMAIN_PROPERTY_MAP: Record<string, string> = {
  $hubspot_company: '$hubspot_company_domain',
  $salesforce_account: '$salesforce_account_website',
  $6signal: '$6Signal_domain',
  $linkedin_company: '$li_domain',
  $g2: '$g2_domain'
};

export const GranularityOptions: string[] = [
  'Timestamp',
  'Hourly',
  'Daily',
  'Weekly',
  'Monthly'
];

export const TIMELINE_VIEW_OPTIONS: string[] = [
  'timeline',
  'birdview',
  'overview'
];

export const DEFAULT_TIMELINE_CONFIG: TimelineConfig = {
  disabled_events: [],
  user_config: {
    table_props: [],
    milestones: []
  },
  account_config: {
    table_props: [],
    milestones: [],
    user_prop: ''
  }
};

export const eventIconsColorMap: Record<
  string,
  { iconColor: string; bgColor: string; borderColor: string }
> = {
  brand: {
    iconColor: '#EE3C3C',
    bgColor: '#FAFAFA',
    borderColor: '#EEEEEE'
  },
  envelope: {
    iconColor: '#FF7875',
    bgColor: '#FFF4F4',
    borderColor: '#FFDEDE'
  },
  handshake: {
    iconColor: '#85A5FF',
    bgColor: '#EFF3FF',
    borderColor: '#D3DEFF'
  },
  phone: {
    iconColor: '#95DE64',
    bgColor: '#F0FFE7',
    borderColor: '#D5F4C1'
  },
  listcheck: {
    iconColor: '#5CDBD3',
    bgColor: '#EBFFFE',
    borderColor: '#C6F6F4'
  },
  'hand-pointer': {
    iconColor: '#FAAD14',
    bgColor: '#FFF3DB',
    borderColor: '#FBE5BA'
  },
  hubspot: {
    iconColor: '#FF7A59',
    bgColor: '#FFE8E2',
    borderColor: '#FED0C5'
  },
  salesforce: {
    iconColor: '#00A1E0',
    bgColor: '#E8F8FF',
    borderColor: '#CDF0FF'
  },
  linkedin: {
    iconColor: '#0A66C2',
    bgColor: '#E6F7FF',
    borderColor: '#91D5FF'
  },
  g2crowd: {
    iconColor: '#FF7A59',
    bgColor: '#FFE8E2',
    borderColor: '#FED0C5'
  },
  window: {
    iconColor: '#FF85C0',
    bgColor: '#FFF0F7',
    borderColor: '#FFD9EB'
  },
  'calendar-star': {
    iconColor: '#B37FEB',
    bgColor: '#F6EDFF',
    borderColor: '#E9D4FF'
  }
};

export const iconColors: string[] = [
  '#85A5FF',
  '#B37FEB',
  '#5CDBD3',
  '#FF9C6E',
  '#FF85C0',
  '#FFC069',
  '#A0D911',
  '#FAAD14'
];

export const ALPHANUMSTR = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ';

export const EngagementTag: Record<string, { bgColor: string; icon: string }> =
  {
    Hot: {
      bgColor: '#FFF1F0',
      icon: 'fire'
    },
    Warm: {
      bgColor: '#FFF7E6',
      icon: 'sun'
    },
    Cool: {
      bgColor: '#F0F5FF',
      icon: 'snowflake'
    },
    Ice: {
      bgColor: '#E6F7FF',
      icon: 'icecube'
    }
  };
