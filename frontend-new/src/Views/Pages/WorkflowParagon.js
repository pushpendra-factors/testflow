import React, { useEffect, useState } from 'react';

import { Text } from 'factorsComponents';
import { paragon } from '@useparagon/connect/dist/src/index';

import LoggedOutScreenHeader from 'Components/GenericComponents/LoggedOutScreenHeader';
import { Button, Card, Col, List, Row, Tag, Typography, message } from 'antd';
import Title from 'antd/lib/skeleton/Title';
import useParagon from 'hooks/useParagon';
import { get, getHostUrl } from 'Utils/request';
import { useSelector } from 'react-redux';
const host = getHostUrl();
const WorkflowParagon = function () {
  const project_id = useSelector((state) => state.global?.active_project?.id);

  const [state, setState] = useState({
    token: ''
  });
  const { user, error, isLoaded } = useParagon(state.token);
  const fetchToken = async () => {
    get(null, `${host}projects/${project_id}/paragon/auth`)
      .then((res) => {
        if (!res?.data) {
          console.error('JWT Token not found');
          return;
        }
        setState((prev) => {
          return {
            ...prev,
            token: res?.data
          };
        });
      })
      .catch((err) => {
        console.error(err);
        message.error('Token not found!');
      });
  };
  useEffect(() => {
    // Authenticate();
    fetchToken();
  }, []);

  return (
    <div>
      <Row style={{ padding: '0 20px', margin: '0 20px' }}>
        <Text type={'title'} level={3}>
          Integrations
        </Text>
        <List
          style={{ width: '100%' }}
          grid={{ gutter: 16, column: 6 }}
          loading={!isLoaded}
          dataSource={paragon.getIntegrationMetadata()}
          renderItem={(integration, index) => {
            const integrationEnabled =
              user.authenticated &&
              user.integrations[integration.type]?.enabled;
            return (
              <List.Item key={integration.type}>
                <Card>
                  <div style={{ display: 'flex', justifyContent: 'left' }}>
                    <img
                      src={integration.icon}
                      style={{ maxWidth: '30px', maxHeight: '30px' }}
                    />{' '}
                    <div style={{ padding: '0 10px', fontSize: '20px' }}>
                      {integration.name}
                    </div>
                  </div>
                  <br />
                  <Button
                    type='primary'
                    onClick={() => paragon.connect(integration.type)}
                  >
                    {' '}
                    {integrationEnabled ? 'Manage' : 'Enable'}
                  </Button>
                </Card>
              </List.Item>
            );
          }}
        />
      </Row>
    </div>
  );
};
export default WorkflowParagon;
