import React from 'react';
import { SVG } from '../../../../components/factorsComponents';

function DurationCol() {
  return (
    <div className='flex items-center justify-between'>
      <div className='text-base' style={{ color: '#8692A3' }}>
        &mdash;
      </div>
      <SVG name='clock' />
      <div className='text-base' style={{ color: '#8692A3', marginTop: '2px' }}>
        &rarr;
      </div>
    </div>
  );
}

export default DurationCol;
