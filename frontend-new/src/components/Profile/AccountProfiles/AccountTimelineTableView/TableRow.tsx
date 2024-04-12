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

function TableRow({ event, eventPropsType = {}, onEventClick }: TableRowProps) {
  const { eventPropNames } = useSelector((state: any) => state.coreQuery);
  const { projectDomainsList, currentProjectSettings } = useSelector(
    (state: any) => state.global
  );

  const timestamp = event?.timestamp
    ? MomentTz(event.timestamp * 1000).format('hh:mm A')
    : '';

  const hasEventProperties =
    currentProjectSettings?.timelines_config?.events_config?.[
      event?.display_name === 'Page View' ? 'PageView' : event?.name
    ]?.length > 0;
  const propertyName =
    currentProjectSettings?.timelines_config?.events_config?.[
      event?.display_name === 'Page View' ? 'PageView' : event?.name
    ]?.[0];

  const renderPropertyName = () =>
    hasEventProperties
      ? `${eventPropNames[propertyName] || PropTextFormat(propertyName)}:`
      : null;

  const renderPropertyValue = () => {
    if (!hasEventProperties) {
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
      ? event?.name
      : Object.entries(event?.properties || {})?.[0]?.[1];

  return (
    <tr
      className={`table-row ${
        event.is_group_user && !hasEventProperties
          ? 'pointer-events-none'
          : 'clickable'
      } cursor-pointer`}
      onClick={onEventClick}
    >
      <td className='timestamp-cell'>{timestamp}</td>
      <td className='event-cell'>
        <EventIcon icon={event.icon} size={24} />
        <TextWithOverflowTooltip
          text={event.display_name || event.name}
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
          title={event.username || event.user_id}
          userID={event.user_id}
          isAnonymous={event.is_anonymous_user}
        />
      </td>
    </tr>
  );
}

export default TableRow;
