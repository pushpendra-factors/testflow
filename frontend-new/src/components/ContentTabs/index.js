import React from 'react';
import { Tabs } from 'antd';
import styles from './index.module.scss';

function ContentTabs({
  onChange, activeKey, tabItems, resultState
}) {
  const { TabPane } = Tabs;

  const loading = !!resultState.find(elem => elem.loading);

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
            <TabPane disabled={loading} className="coreQueryTabPane" tab={getTabTitle(tab)} key={tab.key}>
              {tab.content}
            </TabPane>
          );
        })
      }
    </Tabs>
  );
}

export default ContentTabs;
