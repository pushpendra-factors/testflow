import { Spin } from 'antd';
import React from 'react';
import _ from 'lodash';
import NoDataWithMessage from 'Components/Profile/MyComponents/NoDataWithMessage';
import { groups } from '../../utils';
import SingleTimelineViewTable from './SingleTimelineViewTable';

function UserTimelineSingleView({
  activities = [],
  loading = false,
  propertiesType
}) {
  const groupedActivities = _.groupBy(activities, groups.Daily);

  document.title = 'People - FactorsAI';

  return loading ? (
    <Spin size='large' className='fa-page-loader' />
  ) : activities.length === 0 ? (
    <NoDataWithMessage message='No Events Enabled to Show' />
  ) : (
    <SingleTimelineViewTable
      data={groupedActivities}
      propertiesType={propertiesType}
    />
  );
}

export default UserTimelineSingleView;
