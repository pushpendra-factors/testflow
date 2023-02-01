import { Text } from 'Components/factorsComponents';
import React from 'react';

const BrandInfo = ({ logo, name, description, links }) => {
  return (
    <div className='flex flex-col justify-start gap-1 p-5 pb-0'>
      <div className='flex items-center w-16 h-16'>
        <div className='w-12 h-12'>
          <img
            src={logo}
            alt='logo'
            className='rounded-full border-solid border-2 '
          />
        </div>
      </div>

      <Text
        type={'title'}
        level={6}
        weight={'bold'}
        color='grey-2'
        extraClass='mb-0 mt-1'
      >
        {name}
      </Text>
      <Text type={'paragraph'} mini color='grey' extraClass='mb-0'>
        {description}
      </Text>
      <div className='mt-1 flex items-center'>
        {links?.length > 0
          ? links.map((link) =>
              link?.href && link?.source ? (
                <a href={link.href} target='_blank' rel='noreferrer'>
                  <img src={link.source} alt='link-logo' className='w-4 h-4' />
                </a>
              ) : null
            )
          : null}
      </div>
    </div>
  );
};

export default BrandInfo;
