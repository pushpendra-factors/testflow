import React, { useEffect, useCallback } from 'react';
import { Layout } from 'antd';

function Header(props) {
  const { Header } = Layout;

  const addShadowToHeader = useCallback(() => {
    const scrollTop = (window.pageYOffset !== undefined) ? window.pageYOffset : (document.documentElement || document.body.parentNode || document.body).scrollTop;
    if (scrollTop > 0) {
      document.getElementById('app-header').style.filter = 'drop-shadow(0px 2px 0px rgba(200, 200, 200, 0.25))';
    } else {
      document.getElementById('app-header').style.filter = 'none';
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
      id="app-header"
      className="ant-layout-header--custom bg-white z-20 fixed"
      style={{ width: 'calc(100% - 64px)', padding: 0 }}
    >
      <div className="fa-container">
        {props.children}
      </div>

    </Header>
  );
}

export default Header;
