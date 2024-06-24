import React from 'react';
import { Text } from 'factorsComponents';

const NoDataInTimeRange = ({ message }) => (
  <div className='flex flex-col items-center gap-y-2 text-center fa-no-data--img p-4'>
    <img
      alt='no-data'
      src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/no-data-charts.png'
    />
    <Text type='title' level={8} weight='thin' color='grey' extraClass='m-0'>
      {message}
    </Text>
  </div>
);

export default NoDataInTimeRange;
