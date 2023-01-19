import React, { useState, useEffect } from 'react';
import {
    Row, Col, Button, Spin, Tag, Modal
} from 'antd';
import _, { isEmpty } from 'lodash';
import { Text, SVG, FaErrorComp, FaErrorLog, Number } from 'factorsComponents';


const EventTag = ({ text = 'A', color = 'blue' }) => {
    return (
        <div className={`explain-insight--tag flex justify-center items-center mr-2 ${color ? color : 'blue'}`} style={{ height: '24px', width: '24px' }}><Text type={'title'} level={7} weight={'bold'} color={'white'} extraClass={'m-0'}>{text}</Text> </div>
    )
}
export default EventTag