import React, { useMemo } from 'react';
import MomentTz from 'Components/MomentTz';
import EventIcon from './EventIcon';
import UsernameWithIcon from './UsernameWithIcon';
import { TableRowProps } from './types';
import {
  PropTextFormat,
  convertGroupedPropertiesToUngrouped
} from 'Utils/dataFormatter';
import truncateURL, { isValidURL } from 'Utils/truncateURL';
import { useSelector } from 'react-redux';
import { getPropType, propValueFormat } from 'Components/Profile/utils';
import TextWithOverflowTooltip from 'Components/GenericComponents/TextWithOverflowTooltip';

const TableRow: React.FC<TableRowProps> = ({ event, user, onEventClick }) => {
  const { eventPropertiesV2 } = useSelector((state: any) => state.coreQuery);

  const eventPropertiesModified = useMemo(() => {
    if (!event.event_name) return null;
    const eventProps: any = [];
    if (eventPropertiesV2?.[event.event_name]) {
      convertGroupedPropertiesToUngrouped(
        eventPropertiesV2?.[event.event_name],
        eventProps
      );
    }
    return eventProps;
  }, [event.event_name, eventPropertiesV2]);

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

    const [propertyName, propertyValue] =
      Object.entries(properties || {})[0] || [];
    const value = display_name === 'Page View' ? event_name : propertyValue;

    if (isValidURL(value)) {
      return truncateURL(value);
    }

    const propType = getPropType(eventPropertiesModified, propertyName);
    const formattedValue = propValueFormat(propertyName, value, propType);

    return formattedValue;
  };

  const renderPropValTooltip = () => {
    return event?.display_name === 'Page View'
      ? event?.event_name
      : Object.entries(event?.properties || {})?.[0]?.[1];
  };

  return (
    <tr
      className={`table-row ${isEventClickable ? 'active cursor-pointer' : ''}`}
      onClick={() => isEventClickable && onEventClick(event)}
    >
      <td className='timestamp-cell'>{timestamp}</td>
      <td className='icon-cell'>
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
};

export default TableRow;
