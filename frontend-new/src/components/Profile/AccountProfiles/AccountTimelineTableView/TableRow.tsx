import React from 'react';
import MomentTz from 'Components/MomentTz';
import { Tooltip } from 'antd';
import EventIcon from './EventIcon';
import UsernameWithIcon from './UsernameWithIcon';
import { TableRowProps } from './types';
import { PropTextFormat } from 'Utils/dataFormatter';
import truncateURL, { isValidURL } from 'Utils/truncateURL';

const TableRow: React.FC<TableRowProps> = ({ event, user, onEventClick }) => {
  const timestamp = event?.timestamp
    ? MomentTz(event.timestamp * 1000).format('hh:mm A')
    : '';

  const isEventClickable = Object.keys(event?.properties || {}).length > 0;

  const renderPropertyName = () => {
    if (event?.display_name === 'Page View') {
      return 'Page URL:';
    }
    return isEventClickable
      ? `${PropTextFormat(Object.keys(event?.properties || {})[0])}:`
      : null;
  };

  const renderPropertyValue = () => {
    const { properties, display_name, event_name } = event || {};
    const isEventClickable = Object.keys(properties || {}).length > 0;

    if (!isEventClickable) {
      return null;
    }

    const value =
      display_name === 'Page View'
        ? event_name
        : properties?.[Object.keys(properties || {})[0]];

    const isURL = isValidURL(value);
    const finalValue = isURL ? truncateURL(value) : value;

    return finalValue;
  };

  return (
    <tr className='table-row'>
      <td className='timestamp-cell'>{timestamp}</td>
      <td
        className={`icon-cell ${isEventClickable ? 'cursor-pointer' : ''}`}
        onClick={() => isEventClickable && onEventClick(event)}
      >
        <EventIcon icon={event.icon} size={16} />
        <Tooltip title={event.display_name || event.event_name}>
          <span className='ml-2'>{event.display_name || event.event_name}</span>
        </Tooltip>
      </td>
      <td className='properties-cell'>
        <div className='propkey'>{renderPropertyName()}</div>
        <div className='propvalue'>
          <Tooltip title={renderPropertyValue()}>
            {renderPropertyValue()}
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
};

export default TableRow;
