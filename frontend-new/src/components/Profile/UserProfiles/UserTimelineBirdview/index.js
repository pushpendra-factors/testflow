import React, { useEffect, useState } from 'react';
import { Spin, Tooltip } from 'antd';
import _ from 'lodash';
import { CaretUpOutlined, CaretRightOutlined } from '@ant-design/icons';
import { PropTextFormat } from 'Utils/dataFormatter';
import { useSelector } from 'react-redux';
import NoDataWithMessage from 'Components/Profile/MyComponents/NoDataWithMessage';
import truncateURL from 'Utils/truncateURL';
import { getEventCategory, getIconForCategory, groups } from '../../utils';
import InfoCard from '../../MyComponents/InfoCard';
import { eventIconsColorMap } from 'Components/Profile/constants';

function UserTimelineBirdview({
  activities = [],
  granularity,
  collapse,
  setCollapse,
  loading,
  propertiesType,
  eventNamesMap
}) {
  const [showAll, setShowAll] = useState([]);
  const { userPropNames } = useSelector((state) => state.coreQuery);
  const { projectDomainsList } = useSelector((state) => state.global);

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
          borderColor: `${eventIconsColorMap[eventIcon]?.borderColor}`,
          background: `${eventIconsColorMap[eventIcon]?.bgColor}`
        }}
      >
        <img
          src={`/assets/icons/${eventIcon}.svg`}
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
      Object.keys(event?.properties || {})?.length ||
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
          propertiesType={propertiesType}
          trigger={hoverConditionals ? 'hover' : []}
          icon={
            <img
              src={`/assets/icons/${icon}.svg`}
              alt=''
              height={24}
              width={24}
              loading='lazy'
            />
          }
        >
          <div className='inline-flex gap--6 items-center'>
            <div className='event-name--sm'>
              <Tooltip
                title={eventName}
                trigger={
                  !hoverConditionals && eventName.length >= 30 ? 'hover' : []
                }
              >
                {truncateURL(eventName, projectDomainsList)}
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
        <div className='birdview-events__num' onClick={onClick}>
          {`+${Number(eventsCount - 1)}`}
        </div>
      ) : (
        <div className='birdview-events__num' onClick={onClick}>
          <CaretUpOutlined /> Show Less
        </div>
      )
    ) : null;

  const renderTimeline = (data) =>
    !Object.entries(data).length ? (
      <NoDataWithMessage message='No Activity' />
    ) : (
      <div className='birdview-container bordered-gray--bottom'>
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
                  <td
                    style={{
                      paddingBottom: `${(milestones?.length || 0) * 38}px`
                    }}
                  >
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
                    className='bg-gradient--120px'
                    style={{
                      paddingBottom: `${(milestones?.length || 0) * 38}px`
                    }}
                  >
                    <div
                      className={`birdview-events user-pad ${
                        !showAll[index]
                          ? 'birdview-events--collapsed'
                          : 'birdview-events--expanded'
                      }`}
                    >
                      {eventsList?.map((event) => (
                        <div className='birdview-events__event'>
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
                          <div className='green-stripe opaque' />
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
