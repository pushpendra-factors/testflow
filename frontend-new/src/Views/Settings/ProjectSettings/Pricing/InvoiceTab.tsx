import { Text } from 'Components/factorsComponents';
import { Divider } from 'antd';
import React from 'react';

const InvoiceTab = () => {
  return (
    <div className='py-4'>
      <div className='mb-6'>
        <Text
          type={'title'}
          level={4}
          weight={'bold'}
          extraClass={'m-0 mb-2'}
          color='character-primary'
        >
          Billing and Invoicing
        </Text>
        <Text
          type={'title'}
          level={7}
          extraClass={'m-0'}
          color='character-secondary'
        >
          All invoices that have been generated
        </Text>
        <Divider />
      </div>
    </div>
  );
};

export default InvoiceTab;
