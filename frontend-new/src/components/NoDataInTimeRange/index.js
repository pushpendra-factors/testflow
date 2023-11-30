import React from 'react';
import { Text } from 'factorsComponents';

const NoDataInTimeRange = ({ message }) => {
  return (
    <div className={'flex flex-col items-center fa-no-data--img'}>
      <img
        alt='no-data'
        src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/no-data-charts.png'
        className={'mb-2'}
      />
      <Text
        type={'title'}
        level={8}
        weight={'thin'}
        color={'grey'}
        extraClass={'m-0'}
      >
        {message}
      </Text>
    </div>
  );
};

export default NoDataInTimeRange;
