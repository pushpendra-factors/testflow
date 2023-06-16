import React, { useEffect, useState } from 'react';
import { Spin, Tooltip } from 'antd';
import _ from 'lodash';
import { CaretUpOutlined, CaretRightOutlined } from '@ant-design/icons';
import { SVG } from '../../../factorsComponents';
import InfoCard from '../../MyComponents/InfoCard';
import {
  eventIconsColorMap,
  getEventCategory,
  getIconForCategory,
  groups,
  hoverEvents
} from '../../utils';
import { PropTextFormat } from 'Utils/dataFormatter';
import { useSelector } from 'react-redux';
import NoDataWithMessage from 'Components/Profile/MyComponents/NoDataWithMessage';

function UserTimelineBirdview({
  activities = [],
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

  document.title = 'People - FactorsAI';

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

  const renderIcon = (event) => {
    const eventIcon = eventIconsColorMap[event.icon]
      ? event.icon
      : 'calendar-star';
    return (
      <div
        className='icon'
        style={{
          '--border-color': `${eventIconsColorMap[eventIcon]?.borderColor}`,
          '--bg-color': `${eventIconsColorMap[eventIcon]?.bgColor}`
        }}
      >
        <img
          src={`https://s3.amazonaws.com/www.factors.ai/assets/img/product/Timeline/${eventIcon}.svg`}
          alt=''
          height={16}
          width={16}
          loading='lazy'
        />
      </div>
    );
  };

  const renderInfoCard = (event) => {
    const eventName = event.alias_name
      ? event.alias_name
      : event.display_name !== 'Page View'
      ? PropTextFormat(event.display_name)
      : event.event_name;
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
          eventType={event?.event_type}
          title={event?.alias_name}
          eventSource={event?.display_name}
          eventName={event?.event_name}
          properties={event?.properties || {}}
          trigger={hoverConditionals ? 'hover' : []}
          icon={
            <img
              src={`https://s3.amazonaws.com/www.factors.ai/assets/img/product/Timeline/${icon}.svg`}
              alt=''
              height={24}
              width={24}
              loading='lazy'
            />
          }
          listProperties={listProperties}
        >
          <div className='inline-flex gap--6 items-center'>
            <div className='event-name--sm'>
              <Tooltip
                title={eventName}
                trigger={
                  !hoverConditionals && eventName.length >= 30 ? 'hover' : []
                }
              >
                {eventName}
              </Tooltip>
            </div>
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
      <NoDataWithMessage message={'No Activity'} />
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
              const timelineEvents = events.filter(
                (event) => event.event_type !== 'milestone'
              );
              const eventsList = showAll[index]
                ? timelineEvents
                : timelineEvents.slice(0, 1);
              const milestones = events.filter(
                (event) => event.event_type === 'milestone'
              );
              return (
                <tr>
                  <td className={`pb-${milestones?.length * 8}`}>
                    <div className='timestamp top-40'>{timestamp}</div>
                    {milestones.length ? (
                      <div className='milestone-section'>
                        {milestones.map((milestone) => (
                          <div className='green-stripe'>
                            <div className='text'>
                              {userPropNames[milestone.event_name]
                                ? userPropNames[milestone.event_name]
                                : milestone.event_name}
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
                      {eventsList?.map((event, ind) => (
                        <div key={ind} className='timeline-events__event'>
                          {renderIcon(event)}
                          {renderInfoCard(event)}
                        </div>
                      ))}
                      {renderAdditionalDiv(
                        timelineEvents.length,
                        !showAll[index],
                        () => setShowAllIndex(index, !showAll[index])
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
