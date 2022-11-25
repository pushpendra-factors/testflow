import React, { useEffect, useCallback, useState } from 'react';
import { Layout } from 'antd';
import { SVG, Text } from 'factorsComponents';
import { Link } from 'react-router-dom'; 
import { connect } from 'react-redux'; 

function Header({  
  activeQuery=false
 }) {
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
    // setTimeout(() => document.getElementById("fa-report-container").scrollIntoView({ behavior: 'smooth', block: 'start' }), 2);
    return () => {
      document.removeEventListener('scroll', addShadowToHeader);
    }; 
  }, [addShadowToHeader]);

  return (
    <Header id="app-header" className="ant-layout-header--custom bg-white w-full z-20 fixed px-8 p-0 top-0" > 
      <div className="flex py-4 justify-between items-center">
        <div className="flex items-center items-center">
          <div>
            <Link to="/"><SVG name={'brand'} color="#0B1E39" size={32} /></Link>
          </div> 
        </div>
        <Text type={'title'} level={7} color={'grey'} weight={'bold'} extraClass={`m-0`} >
          {activeQuery ? activeQuery?.title : `Path Analysis`}
        </Text>
        <div className="flex items-center items-center">  
          <Link to="/path-analysis" style={{ color: '#0E2647', fontSize: '14px' }} className='ml-4' ><SVG extraClass="mr-1" name={"close"} size={20} color={'grey'} /></Link> 
        </div>
      </div> 


    </Header>
  );
}

const mapStateToProps = (state) => {
  return {
    activeProject: state.global.active_project
  };
};
export default connect(mapStateToProps, null)(Header);
