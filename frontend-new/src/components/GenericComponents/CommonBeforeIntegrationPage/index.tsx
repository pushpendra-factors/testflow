import React from 'react';
import { Button } from 'antd';
import UnplugImage from '../../../assets/images/unplug.png';
import { Text } from 'Components/factorsComponents';
import { useHistory } from 'react-router-dom';
import { PathUrls } from 'Routes/pathUrls';

const CommonBeforeIntegrationPage = () => {
  const history = useHistory();
  return (
    <div
      className='w-full h-full flex flex-col justify-center items-center'
      style={{ height: 'calc(100vh - 64px)' }}
    >
      <div style={{ height: 165, width: 165, marginTop: -64 }}>
        <img src={UnplugImage} alt='' />
      </div>
      <Text
        type={'title'}
        level={3}
        color='character-title'
        extraClass='m-0 mt-6'
      >
        Website account identification paused
      </Text>
      <Text
        type={'title'}
        level={7}
        color='character-secondary'
        extraClass='m-0 mt-2'
      >
        Account identification is not active on your project. To active, please
        enable ‘Factors Website De-anonymization’ from the integrations page.
      </Text>
      <Button
        className='mt-6'
        type='primary'
        onClick={() => history.push(PathUrls.SettingsIntegration)}
      >
        Go to Integrations
      </Button>
    </div>
  );
};

export default CommonBeforeIntegrationPage;
