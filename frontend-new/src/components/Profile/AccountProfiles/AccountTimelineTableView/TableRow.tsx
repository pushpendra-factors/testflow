import React, { useMemo } from 'react';
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

  const { propertyValue, propertyValueTooltip } = useMemo(() => {
    if (!hasEventProperties) {
      return { propertyValue: null, propertyValueTooltip: null };
    }

    const propValue = event?.properties?.[propertyName];
    const propType = eventPropsType[propertyName];
    const formattedValue = propValueFormat(propertyName, propValue, propType);

    const truncatedValue =
      truncateURL(formattedValue, projectDomainsList) || formattedValue;

    return {
      propertyValue: truncatedValue,
      propertyValueTooltip: formattedValue
    };
  }, [event]);

  return (
    <tr
      className={`table-row ${
        !event.is_group_user && !hasEventProperties
          ? 'pointer-events-none'
          : 'clickable'
      } cursor-pointer`}
      onClick={onEventClick}
    >
      <td className='fixed-cell timestamp-cell'>
        <TextWithOverflowTooltip text={timestamp} />
      </td>
      <td className='fixed-cell event-cell'>
        <EventIcon icon={event.icon} size={24} />
        <TextWithOverflowTooltip
          text={event.display_name || event.name}
          extraClass='text'
        />
      </td>
      <td className='fixed-cell properties-cell'>
        <div className='propkey'>{renderPropertyName()}</div>
        <TextWithOverflowTooltip
          text={propertyValue}
          tooltipText={propertyValueTooltip}
          extraClass='propvalue'
        />
      </td>
      <td className='fixed-cell user-cell'>
        <UsernameWithIcon
          title={event.username || event.user_id}
          userID={event.user_id}
          isAnonymous={event.is_anonymous_user}
          isGroupUser={event.is_group_user}
        />
      </td>
    </tr>
  );
}

export default TableRow;
