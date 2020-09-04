import React from 'react';
import { useSelector } from 'react-redux'
import { Layout } from 'antd';
import EventsInfo from '../CoreQuery/ResultsPage/EventsInfo';

function Header() {
	const { Header, Content } = Layout;

	const globalInfo = useSelector(state => state.global);

	return (
		<Header className="ant-layout-header--custom bg-white" style={{ position: 'fixed', zIndex: 1, width: '100%' }}>
			<div className="fai-global-search--container flex flex-col justify-center items-center">
				<input className="fai--global-search" placeholder={`Lookup factors.ai`} />
			</div>
			{globalInfo.is_funnel_results_visible ? (
				<EventsInfo />
			) : null}
		</Header>
	)
}

export default Header;