import React, { useEffect, useMemo, useState } from 'react';
import { Spin } from 'antd';
import _ from 'lodash';
import { CaretUpOutlined, CaretRightOutlined } from '@ant-design/icons';
import { SVG } from '../../../factorsComponents';
import InfoCard from '../../MyComponents/InfoCard';
import {
  eventIconsColorMap,
  getEventCategory,
  getIconForCategory,
  groups,
  hoverEvents,
  iconMap,
  timestampToString
} from '../../utils';
import { PropTextFormat } from 'Utils/dataFormatter';
import { useSelector } from 'react-redux';

function UserTimelineBirdview({
  activities = [],
  milestones,
  granularity,
  collapse,
  setCollapse,
  loading,
  eventNamesMap,
  listProperties
}) {
  const [showAll, setShowAll] = useState([]);
  const { userPropNames } = useSelector((state) => state.coreQuery);

  const groupedActivities = _.groupBy(activities, groups[granularity]);
  const formattedMilestones = useMemo(() => {
    return Object.entries(milestones || {}).map(([key, value]) => [
      key,
      timestampToString[granularity](value)
    ]);
  }, [milestones, granularity]);

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

  const renderIcon = (event) => (
    <div
      className='icon'
      style={{
        '--border-color': `${
          eventIconsColorMap[event.icon || 'calendar_star'].borderColor
        }`,
        '--bg-color': `${
          eventIconsColorMap[event.icon || 'calendar_star'].bgColor
        }`
      }}
    >
      <img
        src={`https://s3.amazonaws.com/www.factors.ai/assets/img/product/Timeline/${
          iconMap[event.icon] ? iconMap[event.icon] : event.icon
        }.svg`}
        alt=''
        height={16}
        width={16}
        loading='lazy'
      />
    </div>
  );

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
    const category = getEventCategory(event, eventNamesMap);
    const icon = getIconForCategory(category);

    return (
      <div className='tag'>
        <InfoCard
          title={event?.alias_name}
          eventSource={event?.display_name}
          eventName={event?.event_name}
          properties={event?.properties || {}}
          trigger={hoverConditionals ? 'hover' : []}
          icon={
            <img
              src={`https://s3.amazonaws.com/www.factors.ai/assets/img/product/Timeline/${
                iconMap[icon] ? iconMap[icon] : icon
              }.svg`}
              alt=''
              height={24}
              width={24}
              loading='lazy'
            />
          }
          listProperties={listProperties}
        >
          <div className='inline-flex gap--6 items-center'>
            <div className='event-name--sm'>{eventName}</div>
            {hoverConditionals ? (
              <CaretRightOutlined
                style={{ fontSize: '12px', color: '#8692A3' }}
              />
            ) : null}
          </div>
        </InfoCard>
      </div>
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
              <th scope='col' />
            </tr>
          </thead>
          <tbody>
            {Object.entries(data).map(([timestamp, events], index) => {
              const eventsList = showAll[index] ? events : events.slice(0, 1);
              const milestones = formattedMilestones.filter(
                (milestone) => milestone[1] === timestamp
              );
              return (
                <tr>
                  <td>
                    <div className='timestamp top-40'>{timestamp}</div>
                    {milestones.length ? (
                      <div className='milestone-section'>
                        {milestones.map((milestone) => (
                          <div className='green-stripe'>
                            <div className='text'>
                              {userPropNames[milestone[0]]
                                ? userPropNames[milestone[0]]
                                : milestone[0]}
                            </div>
                          </div>
                        ))}
                      </div>
                    ) : null}
                  </td>
                  <td
                    className={`bg-gradient--120px pb-${
                      milestones.length * 10
                    }`}
                  >
                    <div
                      className={`timeline-events user-pad ${
                        !showAll[index]
                          ? 'timeline-events--collapsed'
                          : 'timeline-events--expanded'
                      }`}
                    >
                      {eventsList?.map((event) => (
                        <div className='timeline-events__event'>
                          {renderIcon(event)}
                          {renderInfoCard(event)}
                        </div>
                      ))}
                      {renderAdditionalDiv(events.length, !showAll[index], () =>
                        setShowAllIndex(index, !showAll[index])
                      )}
                    </div>
                    {milestones.length ? (
                      <div className='milestone-section'>
                        {milestones.map((milestone) => (
                          <div className={`green-stripe opaque`} />
                        ))}
                      </div>
                    ) : null}
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
export default UserTimelineBirdview;
