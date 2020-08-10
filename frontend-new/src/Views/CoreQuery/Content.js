import React, { useState, useEffect, useRef } from 'react';
import { Tabs } from 'antd';
import styles from './tabs.module.scss';
import { tabItems } from './utils';
import c3 from 'c3';
import * as d3 from 'd3';


function Content() {
    
    const chartRef = useRef(null);
    const [activeKey, setActiveKey] = useState('1');
    const categories = ['Google', 'Facebook', 'G2', 'Capterra', 'Email'];

    const handleTabChange = (key) => {
        setActiveKey(key);
    }

    const getTabTitle = (tab) => {
        return (
            <div className="flex">{tab.titleIcon}<span>&nbsp;{tab.title}</span></div>
        )
    }


    useEffect(() => {
        c3.generate({
            bindto: chartRef.current,
            data: {
                columns: [
                    ['First Touch', 50, 140, 230, 20, 210],
                    ['Last Touch', 130, 120, 110, 100, 90],
                    ['Max Engaged', 100, 90, 80, 70, 60],
                    ['Min Engaged', 120, 80, 40, 40, 63],
                ],
                type: 'bar',
                colors: {
                    'First Touch': '#014694',
                    'Last Touch': '#009FA1',
                    'Max Engaged': '#96CC6E',
                    'Min Engaged': '#008FA1',
                },
            },
            transition: {
                duration: 1000
            },
            // onrendered: showAnnotation,
            bar: {
                width: {
                    ratio: 0.3,
                },
                zerobased: false,
            },
            axis: {
                x: {
                    type: 'category',
                    categories,
                },
            },
            tooltip: {
                grouped: false,
            },
            grid: {
                y: {
                    show: true,
                },
            },
        });
    }, [activeKey, categories]);

    const { TabPane } = Tabs;

    return (
        <div className="mt-4">
            <Tabs className={styles.coreQueryTabs} activeKey={activeKey} onChange={handleTabChange}>
                {
                    tabItems.map(tab => {
                        return (
                            <TabPane className="coreQueryTabPane" tab={getTabTitle(tab)} key={tab.key}>
                                <div className={styles.coreQueryResultsChart} style={{ margin: '0.25rem' }} ref={chartRef} />
                            </TabPane>
                        )
                    })
                }
            </Tabs>
        </div>

    )
}

export default Content;