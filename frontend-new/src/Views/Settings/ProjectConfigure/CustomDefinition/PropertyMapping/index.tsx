import { SVG, Text } from 'Components/factorsComponents';
import { Button, Col, Row, Spin } from 'antd';
import React, { useEffect, useState } from 'react';

import { bindActionCreators } from 'redux';
import { fetchPropertyMappings } from 'Reducers/settings/middleware';
import { ConnectedProps, connect } from 'react-redux';
import PropertyMappingKPI from '../../PropertySettings/PropertyMappingKPI';
import SavedProperties from '../../PropertySettings/PropertyMappingKPI/savedProperties';

const PropertyMapping = ({
  fetchPropertyMappings,
  activeProject
}: PropertyMappingProps) => {
  const [showForm, setShowForm] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchPropertyMappings(activeProject?.id).then(() => {
      setLoading(false);
    });
  }, []);
  if (loading) {
    return (
      <div className='w-full h-full flex items-center justify-center'>
        <Spin />
      </div>
    );
  }
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

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchPropertyMappings
    },
    dispatch
  );

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project
});
const connector = connect(mapStateToProps, mapDispatchToProps);
type ReduxProps = ConnectedProps<typeof connector>;

type PropertyMappingProps = ReduxProps;

export default connector(PropertyMapping);
