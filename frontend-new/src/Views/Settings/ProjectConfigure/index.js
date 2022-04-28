// import React, { useEffect, useState } from 'react';
// import { Row, Col, Menu } from 'antd';
// import Events from './Events';
// import Properties from './PropertySettings';
// import { fetchSmartEvents } from 'Reducers/events';
// import { connect } from 'react-redux';
// import { useHistory, useLocation } from 'react-router-dom';
// import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
// import { ErrorBoundary } from 'react-error-boundary';
// import ContentGroups from './ContentGroups';
// import Touchpoints from './Touchpoints';
// import CustomKPI from './CustomKPI';
// import ExplainDP from './ExplainDataPoints';
// import FaHeader from '../../../components/FaHeader';

// const MenuTabs = {
//   Events: 'Events',
//   Properties: 'Properties',
//   ContentGroups: 'Content Groups',
//   Touchpoints: 'Touchpoints',
//   CustomKPI: 'Custom KPIs',
//   ExplainDP: 'Top Events and Properties',
// };

// function ProjectConfigure({ activeProject, fetchSmartEvents }) {
//   const [selectedMenu, setSelectedMenu] = useState(MenuTabs.Events);
//   const history = useHistory();
//   let location = useLocation();

//   const handleClick = (e) => {
//     setSelectedMenu(e.key);
//     history.push(`/configure`);

//     if (e.key === MenuTabs.Events) {
//       fetchSmartEvents(activeProject.id);
//     }
//   };

//   return (
//     <>
//       <ErrorBoundary
//         fallback={
//           <FaErrorComp
//             size={'medium'}
//             title={'Settings Error'}
//             subtitle={
//               'We are facing trouble loading project settings. Drop us a message on the in-app chat.'
//             }
//           />
//         }
//         onError={FaErrorLog}
//       >
//         {/* <FaHeader /> */}
//         <div className={'mt-24'}>
//           <Row gutter={[24, 24]} justify='center'>
//             <Col span={20}>
//               <Row gutter={[24, 24]}>
//                 <Col span={24}>
//                   <Text
//                     type={'title'}
//                     level={3}
//                     weight={'bold'}
//                     extraClass={'m-0'}
//                   >
//                     Configure
//                   </Text>
//                   <Text
//                     type={'title'}
//                     level={6}
//                     weight={'regular'}
//                     extraClass={'m-0'}
//                     color={'grey'}
//                   >
//                     {activeProject.name}
//                   </Text>
//                 </Col>
//                 <Col span={24}>
//                   <Row gutter={[24, 24]}>
//                     <Col span={6}>
//                       <Menu
//                         onClick={handleClick}
//                         defaultSelectedKeys={selectedMenu}
//                         mode='inline'
//                         className={'fa-settings--menu'}
//                       >
//                         <Menu.Item key={MenuTabs.Touchpoints}>
//                           {MenuTabs.Touchpoints}
//                         </Menu.Item>
//                         <Menu.Item key={MenuTabs.Events}>
//                           {MenuTabs.Events}
//                         </Menu.Item>
//                         <Menu.Item key={MenuTabs.Properties}>
//                           {MenuTabs.Properties}
//                         </Menu.Item>
//                         <Menu.Item key={MenuTabs.ContentGroups}>
//                           {MenuTabs.ContentGroups}
//                         </Menu.Item>
//                         <Menu.Item key={MenuTabs.CustomKPI}>
//                           {MenuTabs.CustomKPI}
//                         </Menu.Item>
//                         <Menu.Item key={MenuTabs.ExplainDP}>
//                           {MenuTabs.ExplainDP}
//                         </Menu.Item>
//                       </Menu>
//                     </Col>
//                     <Col span={18}>
//                       {selectedMenu == MenuTabs.Touchpoints && <Touchpoints />}
//                       {selectedMenu === MenuTabs.Events && <Events />}
//                       {selectedMenu === MenuTabs.Properties && <Properties />}
//                       {selectedMenu === MenuTabs.ContentGroups && (
//                         <ContentGroups />
//                       )}
//                       {selectedMenu === MenuTabs.CustomKPI && <CustomKPI />}
//                       {selectedMenu === MenuTabs.ExplainDP && <ExplainDP />}
//                     </Col>
//                   </Row>
//                 </Col>
//               </Row>
//             </Col>
//           </Row>
//         </div>
//       </ErrorBoundary>
//     </>
//   );
// }

// const mapStateToProps = (state) => ({
//   activeProject: state.global.active_project,
// });

// export default connect(mapStateToProps, { fetchSmartEvents })(ProjectConfigure);
