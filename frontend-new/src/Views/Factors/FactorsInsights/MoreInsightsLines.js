import React from 'react';
import { Text } from 'factorsComponents';

const MoreInsightsLines = ({ insightCount, onClick }) => {
  return (
          <div className="fa-insight-item--more cursor-pointer" onClick={onClick}>
              <Text type={'title'} weight={'thin'} color={'grey'} align={'center'} level={7} extraClass={'m-0 cursor-pointer'} >{insightCount ? `+${insightCount} More Insights` : '++ More Insights'}</Text>
              <div className={'relative border-bottom--thin-2'}/>
              <div className={'relative border-bottom--thin-2'}/>
          </div>
  );
};

export default MoreInsightsLines;
