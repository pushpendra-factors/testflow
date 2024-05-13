import React from 'react';
import { Text } from 'Components/factorsComponents';
import WidgetCard from './WidgetCard';

function NoDataDashboard() {
  return (
    <div className='flex flex-col justify-center fa-dashboard--no-data-container items-center'>
      <img
        alt='no-data'
        src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/no-data.png'
        className='mb-8'
      />
      <Text type='title' level={5} weight='bold' extraClass='m-0'>
        Add widgets to start monitoring.
      </Text>
      <Text type='title' level={7} color='grey' extraClass='m-0'>
        You can select any of the saved reports and add them to dashboard as
        widgets to monitor your metrics.
      </Text>
    </div>
  );
}

function SortableCards({
  widget,
  durationObj,
  setOldestRefreshTime,
  dashboardRefreshState,
  onDataLoadSuccess
}) {
  if (widget?.length) {
    return (
      <div className='flex flex-wrap'>
        {widget?.map((item) => (
          <WidgetCard
            key={item?.inter_id}
            unit={{ ...item }}
            durationObj={durationObj}
            setOldestRefreshTime={setOldestRefreshTime}
            dashboardRefreshState={dashboardRefreshState}
            onDataLoadSuccess={onDataLoadSuccess}
          />
        ))}
      </div>
    );
  }
  return <NoDataDashboard />;
}

export default SortableCards;
