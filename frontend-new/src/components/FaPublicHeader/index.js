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
      <div className='w-3/4 flex justify-end gap-2 items-center px-6'>
        <ControlledComponent controller={isLoggedIn}>
          <Tooltip
            placement='bottom'
            title={`${
              props?.showShareButton
                ? 'Share'
                : 'Only weekly visitor reports can be shared for easy access'
            }`}
          >
            <Button
              onClick={props?.handleShareClick}
              size='large'
              type='primary'
              icon={
                <SVG
                  name={'link'}
                  color={`${props?.showShareButton ? '#fff' : '#b8b8b8'}`}
                />
              }
              disabled={!props?.showShareButton}
            >
              Share
            </Button>
          </Tooltip>
        </ControlledComponent>

        <ControlledComponent controller={!isLoggedIn}>
          <Button
            onClick={() => history.push('/login')}
            size='large'
            className='mr-2'
          >
            Go to Factors
          </Button>
        </ControlledComponent>
        <ControlledComponent controller={isLoggedIn}>
          <Divider type='vertical' style={{ height: '1.5rem' }} />
          <Button
            onClick={() => history.push('/')}
            size='large'
            icon={<SVG name={'Remove'} color='#8692A3' />}
            type='text'
          ></Button>
        </ControlledComponent>
      </div>
    </Header>
  );
}

export default FaPublicHeader;
