import React from 'react';
import { Layout, Button, Tooltip, Divider } from 'antd';
import styles from './index.module.scss';
import { useHistory } from 'react-router-dom';
import { SVG } from 'factorsComponents';
import useAgentInfo from 'hooks/useAgentInfo';
import ControlledComponent from 'Components/ControlledComponent';

function FaPublicHeader(props) {
  const { Header } = Layout;
  const history = useHistory();
  const { isLoggedIn } = useAgentInfo();
  const handleShareClick = () => {
    props.showDrawer();
  };

  return (
    <Header className={`${styles.faheader}`}>
      <div className='w-1/4 flex items-center'>
        <Button
          onClick={() => history.push('/')}
          className={`${styles.logo}`}
          size='large'
          type='text'
        >
          <img
            alt='brand-logo'
            src='https://s3.amazonaws.com/www.factors.ai/assets/img/header-logo.svg'
          />
        </Button>
      </div>
      <div className='w-3/4 flex justify-end gap-2 items-center'>
        <Tooltip placement='bottom' title='Share'>
          <Button
            onClick={handleShareClick}
            size='large'
            type='primary'
            icon={<SVG name={'link'} color='#fff' />}
          >
            Share
          </Button>
        </Tooltip>
        <ControlledComponent controller={!isLoggedIn}>
          <Button onClick={() => history.push('/login')} size='large'>
            Login
          </Button>
        </ControlledComponent>

        <Button
          size='large'
          type='text'
          icon={<SVG name={'threedot'} />}
          onClick={() => console.log('tripple dots click')}
        ></Button>
        <Divider type='vertical' style={{ height: '1.5rem' }} />
        <Button
          onClick={() => history.push('/login')}
          size='large'
          icon={<SVG name={'Remove'} color='#8692A3' />}
          type='text'
        ></Button>
      </div>
    </Header>
  );
}

export default FaPublicHeader;
