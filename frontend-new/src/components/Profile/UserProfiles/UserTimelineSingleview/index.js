import { Spin } from 'antd';
import React from 'react';
import {
  eventIconsColorMap,
  getEventCategory,
  getIconForCategory,
  groups
} from '../../utils';
import _ from 'lodash';
import EventInfoCard from 'Components/Profile/MyComponents/EventInfoCard';
import NoDataWithMessage from 'Components/Profile/MyComponents/NoDataWithMessage';

function UserTimelineSingleView({
  activities = [],
  loading = false,
  propertiesType,
  eventNamesMap
}) {
  const groupedActivities = _.groupBy(activities, groups['Daily']);

  document.title = 'People' + ' - FactorsAI';

  const SingleTimelineViewTable = ({ data = [] }) => (
    <div className='table-scroll mt-8'>
      <table>
        <thead>
          <tr>
            <th scope='col'>Date</th>
            <th scope='col' />
          </tr>
        </thead>
        <tbody>
          {Object.entries(data).map(([timestamp, events], index) => {
            const timelineEvents = events.filter(
              (event) => event.event_type !== 'milestone'
            );
            const milestones = events.filter(
              (event) => event.event_type === 'milestone'
            );
            if (milestones && !timelineEvents.length) return null;
            return (
              <tr>
                <td>
                  <div className='timestamp top-40'>{timestamp}</div>
                </td>
                <td className={`bg-none pb-${milestones.length * 0}`}>
                  <div className={'user-timeline--events'}>
                    {timelineEvents.map((event) => {
                      const category = getEventCategory(event, eventNamesMap);
                      const sourceIcon = getIconForCategory(category);
                      const eventIcon = eventIconsColorMap[event.icon]
                        ? event.icon
                        : 'calendar-star';
                      return (
                        <EventInfoCard
                          event={event}
                          eventIcon={eventIcon}
                          sourceIcon={sourceIcon}
                          propertiesType={propertiesType}
                        />
                      );
                    })}
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
  ) : activities.length === 0 ? (
    <NoDataWithMessage message={'No Events Enabled to Show'} />
  ) : (
    <SingleTimelineViewTable data={groupedActivities} />
  );
}

export default UserTimelineSingleView;
