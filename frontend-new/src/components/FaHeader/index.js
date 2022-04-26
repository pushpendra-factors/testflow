import React, { useState } from 'react';
import { Layout, Button } from 'antd';
import { SVG } from '../factorsComponents';
import styles from './index.module.scss';
import ProjectModal from '../ProjectModal';
import { useHistory } from 'react-router-dom';

function FaHeader(props) {
  const { Header } = Layout;
  const history = useHistory();

  const onCollapse = () => {
    props.setCollapse(!props.collapse);
  };

  return (
    <Header className={`${styles.faheader}`}>
      <div className={'flex items-center'}>
        <Button
          onClick={onCollapse}
          className='fa-btn--custom mx-2'
          type='text'
        >
          <SVG name={'bars'} />
        </Button>
        <Button
          onClick={() => history.push('/')}
          className={`${styles.logo}`}
          size='large'
          type='text'
        >
          <img src='assets/images/header-logo.svg' />
        </Button>
      </div>
      {props.children}
      <ProjectModal />
    </Header>
  );
}

export default FaHeader;
