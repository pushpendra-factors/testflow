import React from 'react';
import { Input } from 'antd';
import { SearchOutlined } from '@ant-design/icons';
import styles from './index.module.scss';

function SearchBar() {
    return (
        <div>
            <Input size="large" className={styles.app_input} placeholder="large size" prefix={<SearchOutlined />} />
        </div>
    );
}

export default SearchBar;