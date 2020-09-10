import React from 'react';
import { useSelector } from 'react-redux'
import { Layout } from 'antd';
import EventsInfo from '../CoreQuery/ResultsPage/EventsInfo';

function Header(props) {
	const { Header, Content } = Layout;

	const globalInfo = useSelector(state => state.global);

	return (
		<Header className="ant-layout-header--custom bg-white z-20 fixed px-8" style={{ width: 'calc(100% - 64px)' }}>
			{props.children}
		</Header>
	)
}

export default Header;