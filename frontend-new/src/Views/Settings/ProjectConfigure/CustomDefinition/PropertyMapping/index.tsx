import { SVG, Text } from 'Components/factorsComponents';
import { Button, Col, Row } from 'antd';
import React, { useState } from 'react';

import SavedProperties from '../../PropertySettings/PropertyMappingKPI/savedProperties';
import PropertyMappingKPI from '../../PropertySettings/PropertyMappingKPI';

const PropertyMapping = () => {
  const [showForm, setShowForm] = useState(false);
  return (
    <div className='mb-4'>
      {!showForm && (
        <Row>
          <Col span={20}>
            <Text type='title' level={7} color='grey' extraClass='m-0'>
              Align metrics from various platforms, like LinkedIn ads and Google
              ads, using a common property such as 'Campaigns.' Seamlessly
              analyze data across platforms to gain comprehensive insights into
              your marketing efforts.{' '}
              {/* <a
                href='https://help.factors.ai/en/articles/7284109-custom-properties'
                target='_blank'
                rel='noreferrer'
              >
                Learn more
              </a> */}
            </Text>
          </Col>
          <Col span={4}>
            <div className='flex justify-end'>
              <Button
                onClick={() => {
                  setShowForm(true);
                }}
                type='primary'
              >
                <SVG name='plus' size={16} color='white' />
                Add New
              </Button>
            </div>
          </Col>
        </Row>
      )}
      {showForm && <PropertyMappingKPI setShowForm={setShowForm} />}
      {!showForm && <SavedProperties />}
    </div>
  );
};

export default PropertyMapping;
