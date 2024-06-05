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
              Harness the full potential of your advertising data with Custom
              Properties. By associating distinct attributes with your data, you
              gain precise control over configuring and analyzing your ad
              campaigns.
            </Text>
            <Text type='title' level={7} color='grey' extraClass='m-0 mt-2'>
              Customize and tailor your data to align perfectly with your
              business objectives, ensuring optimal insights and enhanced
              advertising optimization.
              <a
                href='https://help.factors.ai/en/articles/7284109-custom-properties'
                target='_blank'
                rel='noreferrer'
              >
                Learn more
              </a>
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
