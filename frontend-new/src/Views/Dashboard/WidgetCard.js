import React from 'react';
import {
	Button
} from 'antd';
import { Text } from '../../components/factorsComponents';
// import { FullscreenOutlined, RightOutlined, LeftOutlined } from '@ant-design/icons';
import { FullscreenOutlined } from '@ant-design/icons';
import CardContent from './CardContent';

function WidgetCard({
	setwidgetModal,
	//   resizeWidth,
	//   widthSize,
	unit,
	dashboard
}) {
	//   const calcWidth = (size) => {
	//     switch (size) {
	//       case 1: return 6;
	//       case 2: return 12;
	//       case 3: return 24;
	//       default: return 12;
	//     }
	//   };

	return (
		<div className={`${unit.title} w-full py-4 px-2`} >
			<div style={{ transition: 'all 0.1s' }} className={'fa-dashboard--widget-card w-full'}>
				{/* <div className={'fa-widget-card--resize-container'}>
					<span className={'fa-widget-card--resize-contents'}>
						{widthSize < 3 && <a onClick={() => resizeWidth(unit.id, '+')}><RightOutlined /></a>}
						{widthSize > 1 && <a onClick={() => resizeWidth(unit.id, '-')}><LeftOutlined /></a>}
					</span>
				</div> */}
				<div className={'fa-widget-card--top flex justify-between items-start'}>
					<div className={'w-full'} >
						<Text ellipsis type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>{unit.title}</Text>
						<Text ellipsis type={'paragraph'} mini color={'grey'} extraClass={'m-0'}>{unit.description}</Text>
						<div className="mt-4">
							<CardContent dashboard={dashboard} unit={unit} />
						</div>
					</div>
					<div className={'flex flex-col justify-start items-start fa-widget-card--top-actions'}>
						<Button size={'large'} onClick={() => setwidgetModal(true)} icon={<FullscreenOutlined />} type="text" />
					</div>
				</div>
			</div>
		</div>
	);
}

export default React.memo(WidgetCard);
