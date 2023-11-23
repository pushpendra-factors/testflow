import React from 'react';
import MomentTz from 'Components/MomentTz';
import { Tooltip } from 'antd';
import {
  hoverEventsColumnProp,
  TimelineHoverPropDisplayNames
} from 'Components/Profile/utils';
import EventIcon from './EventIcon';
import UsernameWithIcon from './UsernameWithIcon';
import { TableRowProps } from './types';
import { PropTextFormat } from 'Utils/dataFormatter';

const TableRow: React.FC<TableRowProps> = ({ event, user, onEventClick }) => (
  <tr className='table-row'>
    <td className='timestamp-cell'>
      {MomentTz(event?.timestamp * 1000).format('hh:mm A')}
    </td>
    <td
      className='icon-cell cursor-pointer'
      onClick={() => onEventClick(event)}
    >
      <EventIcon icon={event.icon} size={16} />
      <Tooltip
        title={event.display_name ? event.display_name : event.event_name}
      >
        <span className='ml-2'>
          {event.display_name ? event.display_name : event.event_name}
        </span>
      </Tooltip>
    </td>
    <td className='properties-cell'>
      <div className='propkey'>
        {event?.display_name === 'Page View'
          ? 'Page URL:'
          : hoverEventsColumnProp?.[event?.event_name]
          ? `${
              TimelineHoverPropDisplayNames[
                hoverEventsColumnProp?.[event?.event_name]
              ] || PropTextFormat(hoverEventsColumnProp?.[event?.event_name])
            }:`
          : null}
      </div>
      <div className='propvalue'>
        <Tooltip
          title={
            event?.display_name === 'Page View'
              ? event?.event_name
              : event?.properties?.[hoverEventsColumnProp?.[event?.event_name]]
          }
        >
          {event?.display_name === 'Page View'
            ? event?.event_name
            : event?.properties?.[hoverEventsColumnProp?.[event?.event_name]]}
        </Tooltip>
      </div>
    </td>
    <td className='user-cell'>
      <UsernameWithIcon
        title={user.title}
        userID={event.id}
        isAnonymous={user.isAnonymous}
      />
    </td>
  </tr>
);

export default TableRow;
