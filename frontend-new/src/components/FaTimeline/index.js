import React, { useEffect, useState } from 'react';
import { Spin } from 'antd';
import _ from 'lodash';
import { CaretUpOutlined, CaretRightOutlined } from '@ant-design/icons';
import { SVG } from '../factorsComponents';
import InfoCard from './InfoCard';
import { getEventCategory, getIconForCategory, getIconForEvent, groups, hoverEvents } from '../Profile/utils';
import { PropTextFormat } from 'Utils/dataFormatter';

function FaTimeline({
  activities = [],
  granularity,
  collapse,
  setCollapse,
  loading,
  eventNamesMap
}) {
  const [showAll, setShowAll] = useState([]);

  const groupedActivities = _.groupBy(activities, groups[granularity]);

  useEffect(() => {
    if (collapse !== undefined) {
      const showAllState = new Array(
        Object.entries(groupedActivities).length
      ).fill(!collapse);
      setShowAll(showAllState);
    }
  }, [collapse]);

  const setShowAllIndex = (ind, flag) => {
    setCollapse(undefined);
    const showAllState = [...showAll];
    showAllState[ind] = flag;
    setShowAll(showAllState);
  };

  const renderInfoCard = (event) => {
    const eventName =
      event.display_name === 'Page View'
        ? event.event_name
        : event?.alias_name || PropTextFormat(event.display_name);
    const hoverConditionals =
      hoverEvents.includes(event.event_name) ||
      event.display_name === 'Page View' ||
      event.event_type === 'CH' ||
      event.event_type === 'CS';
    const category = getEventCategory(event, eventNamesMap)
    const icon = getIconForCategory(category);

    return (
      <InfoCard
        title={event?.alias_name || event.display_name}
        eventName={event?.event_name}
        properties={event?.properties || {}}
        trigger={hoverConditionals ? 'hover' : []}
        icon={<SVG name={icon} size={24} />}
      >
        <div className='inline-flex-gap--6 items-center'>
          <div>
            <SVG name={icon} size={16} />
          </div>
          <div className='event-name--sm'>{eventName}</div>
          {hoverConditionals ? <CaretRightOutlined /> : null}
        </div>
      </InfoCard>
    );
  };

  const renderAdditionalDiv = (eventsCount, collapseState, onClick) =>
    eventsCount > 1 ? (
      collapseState ? (
        <div className='timeline-events__num' onClick={onClick}>
          {`+${Number(eventsCount - 1)}`}
        </div>
      ) : (
        <div className='timeline-events__num' onClick={onClick}>
          <CaretUpOutlined /> Show Less
        </div>
      )
    ) : null;

  const renderTimeline = (data) =>
    !Object.entries(data).length ? (
      <div className='ant-empty ant-empty-normal'>
        <div className='ant-empty-image'>
          <SVG name='nodata' />
        </div>
        <div className='ant-empty-description'>No Activity</div>
      </div>
    ) : (
      <div className='table-scroll'>
        <table>
          <thead>
            <tr>
              <th scope='col'>Date and Time</th>
              <th scope='col'></th>
            </tr>
          </thead>
          <tbody>
            {Object.entries(data).map(([timestamp, events], index) => {
              const eventsList = showAll[index] ? events : events.slice(0, 1);
              return (
                <tr>
                  <td>
                    <div className='top-40'>{timestamp}</div>
                  </td>
                  <td className='bg-gradient--120px'>
                    <div
                      className={`timeline-events single-user--padding ${
                        !showAll[index]
                          ? 'timeline-events--collapsed'
                          : 'timeline-events--expanded'
                      }`}
                    >
                      {eventsList?.map((event) => (
                        <div className='timeline-events__event'>
                          {renderInfoCard(event)}
                        </div>
                      ))}
                      {renderAdditionalDiv(events.length, !showAll[index], () =>
                        setShowAllIndex(index, !showAll[index])
                      )}
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    );

  return loading ? (
    <Spin size='large' className='fa-page-loader' />
  ) : (
    renderTimeline(groupedActivities)
  );
}
export default FaTimeline;
