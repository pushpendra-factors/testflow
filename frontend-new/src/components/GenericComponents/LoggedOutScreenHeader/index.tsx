import { SVG } from 'Components/factorsComponents';
import React from 'react';
import { Link } from 'react-router-dom';
import HelpButton from '../HelpButton';

const LoggedOutScreenHeader = ({ helpMessage }: { helpMessage?: string }) => {
  return (
    <div
      className='flex justify-between items-center px-10'
      style={{ height: 64 }}
    >
      <Link
        className='flex items-center font-semibold gap-2'
        target='_blank'
        to={{
          pathname: 'https://www.factors.ai'
        }}
      >
        <SVG name={'BrandFull'} width={140} color='white' />
      </Link>

      <HelpButton helpMessage={helpMessage} />
    </div>
  );
};

export default LoggedOutScreenHeader;
