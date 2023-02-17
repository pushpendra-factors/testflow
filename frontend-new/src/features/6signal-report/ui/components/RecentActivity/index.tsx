import React from 'react';
import { Text } from 'Components/factorsComponents';

const RecentActivity = ({ recentActivities }: RecentActivityProps) => {
  return (
    <div className='flex flex-col justify-start gap-1 px-5 pb-0'>
      <div className='flex flex-col justify-start gap-1'>
        {recentActivities &&
          recentActivities?.length > 0 &&
          recentActivities.map((activity) =>
            activity ? (
              <Text type={'paragraph'} mini color='grey' extraClass='mb-0'>
                {activity}
              </Text>
            ) : null
          )}
      </div>
    </div>
  );
};

type RecentActivityProps = {
  recentActivities: string[];
};

export default RecentActivity;
