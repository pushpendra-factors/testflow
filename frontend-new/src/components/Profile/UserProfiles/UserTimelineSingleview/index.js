import { Spin } from 'antd';
import React, { useMemo } from 'react';
import {
  getEventCategory,
  getIconForCategory,
  groups,
  timestampToString
} from '../../utils';
import _ from 'lodash';
import EventInfoCard from 'Components/Profile/MyComponents/EventInfoCard';
import NoDataWithMessage from 'Components/Profile/MyComponents/NoDataWithMessage';

function AccountTimelineSingleView({
  activities = [],
  milestones,
  loading = false,
  eventNamesMap,
  listProperties
}) {
  const groupedActivities = _.groupBy(activities, groups['Daily']);
  const formattedMilestones = useMemo(() => {
    return Object.entries(milestones || {}).map(([key, value]) => [
      key,
      timestampToString['Daily'](value)
    ]);
  }, [milestones]);

  const SingleTimelineViewTable = ({ data = [] }) => (
    <div className='table-scroll'>
      <table>
        <thead>
          <tr>
            <th scope='col'>Date</th>
            <th scope='col' />
          </tr>
        </thead>
        <tbody>
          {Object.entries(data).map(([timestamp, events], index) => {
            const milestones = formattedMilestones.filter(
              (milestone) => milestone[1] === timestamp
            );
            return (
              <tr>
                <td>
                  <div className='timestamp top-40'>{timestamp}</div>
                  {/* {milestones.length ? (
                    <div className='milestone-section'>
                      {milestones.map((milestone) => (
                        <div className='green-stripe'>
                          <div className='text'>{milestone[0]}</div>
                        </div>
                      ))}
                    </div>
                  ) : null} */}
                </td>
                <td className={`bg-none pb-${milestones.length * 0}`}>
                  <div class='user-timeline--events'>
                    {events.map((event) => {
                      const category = getEventCategory(event, eventNamesMap);
                      const sourceIcon = getIconForCategory(category);
                      const eventIcon = event.icon
                        ? event.icon
                        : 'calendar_star';
                      return (
                        <EventInfoCard
                          event={event}
                          eventIcon={eventIcon}
                          sourceIcon={sourceIcon}
                          listProperties={listProperties}
                        />
                      );
                    })}
                  </div>
                  {/* {milestones.length ? (
                    <div className='milestone-section'>
                      {milestones.map((milestone) => (
                        <div className={`green-stripe opaque`} />
                      ))}
                    </div>
                  ) : null} */}
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

export default AccountTimelineSingleView;
