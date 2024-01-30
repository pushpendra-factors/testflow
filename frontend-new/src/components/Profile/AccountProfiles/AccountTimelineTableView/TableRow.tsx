import React from 'react';
import MomentTz from 'Components/MomentTz';
import { PropTextFormat } from 'Utils/dataFormatter';
import { useSelector } from 'react-redux';
import { propValueFormat } from 'Components/Profile/utils';
import TextWithOverflowTooltip from 'Components/GenericComponents/TextWithOverflowTooltip';
import truncateURL from 'Utils/truncateURL';
import { TableRowProps } from 'Components/Profile/types';
import UsernameWithIcon from './UsernameWithIcon';
import EventIcon from './EventIcon';

function TableRow({
  event,
  eventPropsType = {},
  user,
  onEventClick
}: TableRowProps) {
  const { eventPropNames } = useSelector((state: any) => state.coreQuery);
  const { projectDomainsList } = useSelector((state: any) => state.global);

  const timestamp = event?.timestamp
    ? MomentTz(event.timestamp * 1000).format('hh:mm A')
    : '';

  const isEventClickable = Object.keys(event?.properties || {}).length > 0;

  const renderPropertyName = () => {
    if (event?.display_name === 'Page View') {
      return 'Page URL:';
    }
    return isEventClickable
      ? `${
          eventPropNames[Object.keys(event?.properties || {})[0]] ||
          PropTextFormat(Object.keys(event?.properties || {})[0])
        }:`
      : null;
  };

  const renderPropertyValue = () => {
    const { properties, display_name, event_name } = event || {};

    if (!isEventClickable) {
      return null;
    }

    const [propertyName, propertyValue] =
      Object.entries(properties || {})[0] || [];
    const value = display_name === 'Page View' ? event_name : propertyValue;

    const propType = eventPropsType[propertyName];
    const formattedValue = propValueFormat(propertyName, value, propType);

    return truncateURL(formattedValue, projectDomainsList) || formattedValue;
  };

  const renderPropValTooltip = () =>
    event?.display_name === 'Page View'
      ? event?.event_name
      : Object.entries(event?.properties || {})?.[0]?.[1];

  return (
    <tr
      className={`table-row ${
        isEventClickable ? 'clickable cursor-pointer' : ''
      }`}
      onClick={() => isEventClickable && onEventClick(event)}
    >
      <td className='timestamp-cell'>{timestamp}</td>
      <td className='event-cell'>
        <EventIcon icon={event.icon} size={24} />
        <TextWithOverflowTooltip
          text={event.display_name || event.event_name}
          extraClass='ml-2'
        />
      </td>
      <td className='properties-cell'>
        <div className='propkey'>{renderPropertyName()}</div>
        <TextWithOverflowTooltip
          text={renderPropertyValue()}
          tooltipText={renderPropValTooltip()}
          extraClass='propvalue'
        />
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
}

export default TableRow;
