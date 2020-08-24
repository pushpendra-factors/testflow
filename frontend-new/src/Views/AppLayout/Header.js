import React from 'react'; 
import { Layout } from 'antd'; 

function Header() {
    const { Header, Content } = Layout;

    return (  
            <Header className="ant-layout-header--custom" style={{ position: 'fixed', zIndex: 1, width: '100%' }}> 
                    <div className="fai-global-search--container flex flex-col justify-center items-center">
                            <input className="fai--global-search" placeholder={`Lookup factors.ai`} /> 
                    </div>  
            </Header> 
    )
}

export default Header;