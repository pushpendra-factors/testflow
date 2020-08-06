import React, { useCallback, useState } from 'react';
import { Button, Popover } from 'antd';
import styles from './index.module.scss';
import { Tabs } from 'antd';

function SaveQueryButton() {

    const [saveQueryVisible, setSaveQueryVisible] = useState(false);


    const saveQueryToggle = useCallback(() => {
        setSaveQueryVisible(currState => {
            return !currState;
        })
    }, []);

    const getPopOverTitle = () => {
        return (
            <div className={`p-3 w-64 flex items-center justify-between ${styles.popoverHeading}`}>
                <span>Save as</span>
                <svg onClick={saveQueryToggle} className="cursor-pointer" width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path fillRule="evenodd" clipRule="evenodd" d="M18.7071 6.70711C19.0976 6.31658 19.0976 5.68342 18.7071 5.29289C18.3166 4.90237 17.6834 4.90237 17.2929 5.29289L12 10.5858L6.70711 5.29289C6.31658 4.90237 5.68342 4.90237 5.29289 5.29289C4.90237 5.68342 4.90237 6.31658 5.29289 6.70711L10.5858 12L5.29289 17.2929C4.90237 17.6834 4.90237 18.3166 5.29289 18.7071C5.68342 19.0976 6.31658 19.0976 6.70711 18.7071L12 13.4142L17.2929 18.7071C17.6834 19.0976 18.3166 19.0976 18.7071 18.7071C19.0976 18.3166 19.0976 17.6834 18.7071 17.2929L13.4142 12L18.7071 6.70711Z" fill="grey" />
                </svg>
            </div>
        )
    }

    const getPopOverContent = () => {
        return (
            <div>bannat</div>
        )
    }


    return (
        <Popover visible={saveQueryVisible} placement="bottomLeft" title={getPopOverTitle} content={getPopOverContent} trigger="click">
            <Button onClick={saveQueryToggle} size="large" className={styles.btn} type="primary">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <mask id="path-1-inside-1" fill="white">
                        <rect x="4" y="4.03073" width="7" height="7.96927" rx="1" />
                    </mask>
                    <rect x="4" y="4.03073" width="7" height="7.96927" rx="1" stroke="white" strokeWidth="4" mask="url(#path-1-inside-1)" />
                    <mask id="path-2-inside-2" fill="white">
                        <rect x="4" y="14" width="7" height="6" rx="1" />
                    </mask>
                    <rect x="4" y="14" width="7" height="6" rx="1" stroke="white" strokeWidth="4" mask="url(#path-2-inside-2)" />
                    <path d="M19.2662 16H13.9524" stroke="white" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                    <path d="M16.589 18.664L16.5891 13.4958" stroke="white" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                    <mask id="path-5-inside-3" fill="white">
                        <rect x="13.093" y="4.03073" width="6.90698" height="5.96927" rx="1" />
                    </mask>
                    <rect x="13.093" y="4.03073" width="6.90698" height="5.96927" rx="1" stroke="white" strokeWidth="4" mask="url(#path-5-inside-3)" />
                </svg>
                &nbsp;Save query as
            </Button>
        </Popover>
    )
}

export default SaveQueryButton;