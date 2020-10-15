import React, { useState, useEffect } from 'react';
import {
  Row, Col, Tabs, Modal, Button
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import WidgetCard from './WidgetCard';
import { ReactSortable } from 'react-sortablejs';
const { TabPane } = Tabs;

const widgetCardCollection = [
  {
    id: 1, type: 1, size: 1, title: 'item 1'
  },
  {
    id: 2, type: 2, size: 1, title: 'item 2'
  },
  {
    id: 3, type: 3, size: 1, title: 'item 3'
  },
  {
    id: 4, type: 1, size: 1, title: 'item 4'
  },
  {
    id: 5, type: 2, size: 2, title: 'item 5'
  },
  {
    id: 6, type: 1, size: 1, title: 'item 6'
  },
  {
    id: 7, type: 3, size: 1, title: 'item 7'
  },
  {
    id: 8, type: 2, size: 1, title: 'item 8'
  },
  {
    id: 9, type: 3, size: 1, title: 'item 9'
  },
  {
    id: 10, type: 1, size: 2, title: 'item 10'
  },
  {
    id: 11, type: 2, size: 3, title: 'item 11'
  }
];

function ProjectTabs({ setaddDashboardModal }) {
  const [widgetModal, setwidgetModal] = useState(false);
  const [widgets, setWidgets] = useState(widgetCardCollection);

  const onDrop = (newState) => {
    setWidgets(newState);
  };

  const resizeWidth = (index, operator) => {
    const newArray = [...widgets];
    let newSize = 1;
    const currentWidth = newArray[index].size;
    if (operator === '+') {
      if (currentWidth !== 3) {
        newSize = currentWidth + 1;
      } else {
        newSize = 3;
      }
    } else {
      if (currentWidth !== 0) {
        (currentWidth - 1 === 0) ? newSize = 1 : newSize = currentWidth - 1;
      } else {
        newSize = 1;
      }
    }
    newArray[index] = { ...newArray[index], size: newSize };
    setWidgets(newArray);
  };

  const operations = <>
  <Button type="text" size={'small'} onClick={() => setaddDashboardModal(true)}><SVG name="plus" color={'grey'}/></Button>
  <Button type="text" size={'small'}><SVG name="edit" color={'grey'}/></Button>
  </>;

  useEffect(() => {
    console.log('widgets', widgets);
  });

  return (<>

        <Row className={'mt-2'}>
          <Col span={24}>
             <Tabs defaultActiveKey="1"
              className={'fa-tabs--dashboard'}
              tabBarExtraContent={operations}
              >
                <TabPane tab="My Dashboard" key="1">
                   <div className={'fa-container mt-6'}>
                   <ReactSortable className={'ant-row'} list={widgets} setList={onDrop}>
                      {widgets.map((item, index) => {
                        return (
                          <WidgetCard widthSize={item.size} key={index} index={index} resizeWidth={resizeWidth} title={item.title} id={item.type}/>
                        );
                      })}
                  </ReactSortable>
                   </div>
                </TabPane>
                <TabPane tab="Paid Marketing" key="2">
                <div className={'fa-container mt-6'}>
                  <ReactSortable className={'ant-row'} list={widgets} setList={onDrop}>
                      {widgets.map((item, index) => {
                        return (
                          <WidgetCard widthSize={item.size} key={index} index={index} resizeWidth={resizeWidth} title={item.title} id={item.type}/>
                        );
                      })}
                  </ReactSortable>
                </div>
                </TabPane>
                <TabPane tab="Campaigns" key="3">
                <div className={'fa-container mt-6'}>
                    <ReactSortable className={'ant-row'} list={widgets} setList={onDrop}>
                        {widgets.map((item, index) => {
                          return (
                            <WidgetCard widthSize={item.size} key={index} index={index} resizeWidth={resizeWidth} title={item.title} id={item.type}/>
                          );
                        })}
                    </ReactSortable>
                   </div>
                </TabPane>
              </Tabs>
          </Col>
        </Row>

    <Modal
        title={null}
        visible={widgetModal}
        footer={null}
        centered={false}
        zIndex={1015}
        mask={false}
        onCancel={() => setwidgetModal(false)}
        // closable={false}
        className={'fa-modal--full-width'}
        >
            <div className={'py-10 flex justify-center'}>
                <div className={'fa-container'}>
                    <Text type={'title'} level={5} weight={'bold'} size={'grey'} extraClass={'m-0'}>Full width Modal</Text>
                    <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0'}>Core Query results page comes here..</Text>
                </div>
            </div>
    </Modal>

  </>);
}

export default ProjectTabs;
