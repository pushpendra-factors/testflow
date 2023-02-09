import React from 'react';
import { Text } from 'Components/factorsComponents';

const Properties = ({ properties }) => {
  return (
    <div className='flex flex-col justify-start gap-1 px-5 pb-0'>
      <div className='flex flex-col justify-start gap-4'>
        {properties &&
          properties?.length > 0 &&
          properties.map((property) => {
            if (!property?.name || !property?.value) return null;
            return (
              <div key={property.name}>
                <Text type={'paragraph'} mini color='grey' extraClass='mb-0'>
                  {property.name}:
                </Text>
                <Text type={'paragraph'} color='grey-2' extraClass='mb-0'>
                  {property.value}
                </Text>
              </div>
            );
          })}
      </div>
    </div>
  );
};

export default Properties;
