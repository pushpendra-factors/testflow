import React from 'react';
import { useSelector } from 'react-redux'
import { Layout } from 'antd';

function Header(props) {
	const { Header, Content } = Layout;

	const globalInfo = useSelector(state => state.global);

	return (
		<Header
			className="ant-layout-header--custom bg-white z-20 fixed"
			style={{ width: 'calc(100% - 64px)', filter: 'drop-shadow(0px 2px 0px rgba(200, 200, 200, 0.25))' }}
		>
			<div className="fa-container">
				{props.children}
			</div>

		</Header>
	)
}

export default Header;