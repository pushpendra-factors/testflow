import { SVG } from 'Components/factorsComponents';
import React from 'react';

const NoDataWithMessage = ({ message }) => (
  <div className='ant-empty ant-empty-normal'>
    <div className='ant-empty-image'>
      <SVG name='nodata' />
    </div>
    <div className='ant-empty-description'>{message}</div>
  </div>
);

export default NoDataWithMessage;
