import React from 'react';
import { useSelector } from 'react-redux'
import { Layout } from 'antd';

function Header(props) {
	const { Header, Content } = Layout;

	const globalInfo = useSelector(state => state.global);

	return (
		<Header
			className="ant-layout-header--custom bg-white z-20 fixed"
			style={{ width: '1280px', left: '50%', marginLeft: '-608px', paddingLeft: '40px', paddingRight: '40px' }}
		>
			{props.children}
		</Header>
	)
}

export default Header;