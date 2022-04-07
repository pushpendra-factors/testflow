import React from 'react';
import ColumnsHeading from './ColumnsHeading';

const ColumnsWrapper = ({ children, heading }) => {
  return (
    <div className='flex flex-col gap-y-5'>
      <ColumnsHeading heading={heading} />
      {children}
    </div>
  );
};

export default ColumnsWrapper;
