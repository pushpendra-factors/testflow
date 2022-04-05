import React, { useEffect, useCallback } from 'react';
import { Layout } from 'antd';
import { useSelector } from 'react-redux';

function Header(props) {
  const { Header } = Layout;

  const { show_analytics_result } = useSelector((state) => state.coreQuery);

  let headerWidth = 'calc(100% - 64px)';

  if (show_analytics_result) {
    headerWidth = '100%';
  }

  const addShadowToHeader = useCallback(() => {
    const scrollTop =
      window.pageYOffset !== undefined
        ? window.pageYOffset
        : (
            document.documentElement ||
            document.body.parentNode ||
            document.body
          ).scrollTop;
    if (scrollTop > 0) {
      document.getElementById('app-header').style.borderBottom =
        'thin solid #E7E9ED';
    } else {
      document.getElementById('app-header').style.borderBottom = 'none';
    }
  }, []);

  useEffect(() => {
    document.addEventListener('scroll', addShadowToHeader);
    return () => {
      document.removeEventListener('scroll', addShadowToHeader);
    };
  }, [addShadowToHeader]);

  return (
    <Header
      id='app-header'
      className='ant-layout-header--custom bg-white z-max fixed'
      style={{ width: headerWidth, padding: 0 }}
    >
      <div className='fa-container'>{props.children}</div>
    </Header>
  );
}

export default Header;
