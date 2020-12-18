import React, { useEffect, useCallback } from 'react';
import { Layout } from 'antd';
import { SVG } from 'factorsComponents';
import { Link } from 'react-router-dom';

function Header() {
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
    <Header id="app-header" className="ant-layout-header--custom bg-white w-full z-20 fixed p-0 top-0" >

         <div className="flex py-4 justify-between items-center">
                <div className="leading-4">
                    <div className="flex items-center items-center">

                        <div>
                            <Link to="/"><SVG name={'brand'} color="#0B1E39" size={32} /></Link>
                        </div>
                        <div style={{ color: '#0E2647', opacity: 0.56, fontSize: '14px' }} className="font-bold leading-5 ml-2">  <Link to="/explain" style={{ color: '#0E2647', fontSize: '14px' }} >Factors</Link> / New Goal</div>
                    </div>
                </div>
        </div>

    </Header>
  );
}

export default Header;
