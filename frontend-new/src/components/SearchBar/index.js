import React from 'react';
import { Input } from 'antd';
import { SearchOutlined } from '@ant-design/icons';

function SearchBar() {
    return (
        <Input size="large" placeholder="large size" prefix={<SearchOutlined />} />
    );
}

export default SearchBar;