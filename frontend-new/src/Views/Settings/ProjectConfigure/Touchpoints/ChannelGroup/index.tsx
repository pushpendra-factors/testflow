import { SVG, Text } from 'Components/factorsComponents';
import { Button, Col, Row, Tooltip } from 'antd';
import React, { useState } from 'react';
import useAgentInfo from 'hooks/useAgentInfo';
import DCGTable from '../../PropertySettings/DCG/DCGTable';
import PropertyValueModalDCG from '../../PropertySettings/DCG/PropetyValueModalDCG';

const ChannelGroup = () => {
  const [showDCGForm, setShowDCGForm] = useState(false);
  const [isModalVisible, setShowModalVisible] = useState(false);
  const [editProperty, setEditProperty] = useState(null);

  const { isAdmin } = useAgentInfo();
  const enableEdit = !isAdmin;
  return (
    <div className='mb-4'>
      {!showDCGForm && (
        <Row>
          <Col span={20}>
            <Text type='title' level={7} extraClass='m-0' color='grey'>
              Assign the default channel to all website sessions based on
              conditions. These rules are checked chronologically from top of
              bottom to assign the channel from which a user came from.
            </Text>
          </Col>
          <Col span={4}>
            <div className='flex justify-end'>
              <Tooltip
                placement='top'
                trigger='hover'
                title={enableEdit ? 'Only Admin can edit' : null}
              >
                <Button
                  disabled={enableEdit}
                  onClick={() => {
                    setShowDCGForm(true);
                    setShowModalVisible(true);
                  }}
                  type='primary'
                >
                  <SVG name='plus' size={16} color='white' />
                  Add New
                </Button>
              </Tooltip>
            </div>
          </Col>
        </Row>
      )}
      {!showDCGForm && (
        <DCGTable
          setEditProperty={setEditProperty}
          setShowModalVisible={setShowModalVisible}
          enableEdit={enableEdit}
        />
      )}
      <PropertyValueModalDCG
        isModalVisible={isModalVisible}
        setShowModalVisible={setShowModalVisible}
        setShowDCGForm={setShowDCGForm}
        editProperty={editProperty}
        setEditProperty={setEditProperty}
      />
    </div>
  );
};

export default ChannelGroup;
