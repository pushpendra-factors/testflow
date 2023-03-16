/* eslint-disable */
import React from 'react';
import { Breadcrumb, Row, Col, Divider, Tag } from 'antd';
import { Number, Text } from 'factorsComponents';

const CheckBoxLib = () => {
  const NumberList = [
    {
      number: 100.1234,
      suffix: '',
      prefix: '',
      shortHand: false,
      info: 'Decimal'
    },
    { number: 1000000, suffix: '', prefix: '', shortHand: false, info: false },
    {
      number: 12345,
      suffix: '',
      prefix: '$',
      shortHand: false,
      info: 'Prefix $'
    },
    {
      number: 123456,
      suffix: '$',
      prefix: '',
      shortHand: false,
      info: 'Suffix $'
    },
    {
      number: 123456.123,
      suffix: '%',
      prefix: '',
      shortHand: false,
      info: 'Suffix % + Decimal'
    },
    {
      number: 123456,
      suffix: '#',
      prefix: '',
      shortHand: false,
      info: 'Prefix #'
    },
    {
      number: 120000,
      suffix: '',
      prefix: '',
      shortHand: true,
      info: 'shortHand'
    },
    {
      number: 1230000,
      suffix: '',
      prefix: '',
      shortHand: true,
      info: 'shortHand'
    },
    {
      number: 123450000,
      suffix: '',
      prefix: '',
      shortHand: true,
      info: 'shortHand'
    },
    {
      number: 1234560000,
      suffix: '',
      prefix: '',
      shortHand: true,
      info: 'shortHand'
    },
    {
      number: 123456780000,
      suffix: '',
      prefix: '',
      shortHand: true,
      info: 'shortHand + Decimal'
    }
  ];

  return (
    <>
      <div className='mt-20 mb-8'>
        <Divider orientation='left'>
          <Breadcrumb>
            <Breadcrumb.Item> Components </Breadcrumb.Item>
            <Breadcrumb.Item> Number </Breadcrumb.Item>
          </Breadcrumb>
        </Divider>
      </div>

      <Row>
        <Col span={20}>
        <CodeBlock preClassName={'my-4 fa-code-block'} codeClassName={'fa-code-code-block'} codeContent={`import {Number} from 'factorsComponents';

//Props
<Number 
  prefix={''} //Optional | 'String' e.g: %,#,*
  shortHand={false} //Optional | 'Boolean'
  suffix={''} //Optional | 'String' e.g: %,#,*
  number={number} //Requried
/>
`}></CodeBlock>
      
        </Col>
      </Row>

      <Row className={'mb-8'}>
        <Col span={20}>
          {NumberList.map((item, index) => {
            return (
              <Col span={20} key={index}>
                <div className='flex mb-2 items-center'>
                  <Text
                    type={'title'}
                    level={7}
                    extraClass={'m-0 '}
                  >{`${item.number}  => `}</Text>
                  <Text
                    type={'title'}
                    level={6}
                    extraClass={'m-0 ml-1 mr-2'}
                    weight={'bold'}
                  >
                    <Number
                      prefix={item.prefix}
                      shortHand={item.shortHand}
                      suffix={item.suffix}
                      number={item.number}
                    />
                  </Text>
                  {item.info && (
                    <Tag
                      color='green'
                      style={{
                        borderRadius: '4px',
                        padding: '0 5px',
                        fontSize: '10px'
                      }}
                    >
                      {item.info}
                    </Tag>
                  )}
                </div>
              </Col>
            );
          })}
        </Col>
      </Row>
    </>
  );
};

export default CheckBoxLib;
