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
  const { projectDomainsList, currentProjectSettings } = useSelector(
    (state: any) => state.global
  );

  const timestamp = event?.timestamp
    ? MomentTz(event.timestamp * 1000).format('hh:mm A')
    : '';

  const isEventClickable =
    currentProjectSettings?.timelines_config?.events_config?.[
      event?.display_name === 'Page View' ? 'PageView' : event?.event_name
    ]?.length > 0;
  const propertyName =
    currentProjectSettings?.timelines_config?.events_config?.[
      event?.display_name === 'Page View' ? 'PageView' : event?.event_name
    ]?.[0];

  const renderPropertyName = () =>
    isEventClickable
      ? `${eventPropNames[propertyName] || PropTextFormat(propertyName)}:`
      : null;

  const renderPropertyValue = () => {
    if (!isEventClickable) {
      return null;
    }

    const propertyValue = event?.properties?.[propertyName];
    const propType = eventPropsType[propertyName];
    const formattedValue = propValueFormat(
      propertyName,
      propertyValue,
      propType
    );

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
