import React, { useState } from 'react';
import { Drawer, Button } from 'antd';

import styles from './index.module.scss';

function QueryComposer() {
    const [visible, setVisible] = useState(true);
    const showDrawer = () => {
        setVisible(true);
    };
    const onClose = () => {
        setVisible(false);
    };

    const title = () => {
        return (<div className="composer_title">
            <span>Event Analysis</span>
        </div>)
    }

    return(
        <Drawer
        title={title()}
        placement="left"
        closable={true}
        visible={false}
        onClose={onClose}
        getContainer={false}
        width={"600px"}
        className={styles.query_composer}
        style={{ position: 'absolute'}}
      >

      </Drawer>
    )
}

export default QueryComposer;