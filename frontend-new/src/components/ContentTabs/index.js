import React from 'react';
import { Tabs } from 'antd';
import styles from './index.module.scss';

function ContentTabs({ onChange, activeKey, tabItems }) {
  const { TabPane } = Tabs;

  const getTabTitle = (tab) => {
    return (
      <div className="flex items-center">{tab.titleIcon}<span>&nbsp;{tab.title}</span></div>
    );
  };

  return (
    <Tabs className={styles.contentTabs} activeKey={activeKey} onChange={onChange}>
      {
        tabItems.map(tab => {
          return (
            <TabPane className="coreQueryTabPane" tab={getTabTitle(tab)} key={tab.key}>
              {tab.content}
            </TabPane>
          );
        })
      }
    </Tabs>
  );
}

export default ContentTabs;
