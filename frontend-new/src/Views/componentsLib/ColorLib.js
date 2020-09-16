import React from 'react';
import {
  Layout, Breadcrumb, Row, Col, Divider
} from 'antd';
import { Text } from 'factorsComponents';

function ColorLib() {
  const { Content } = Layout;
  return (
    <>
      <div className="mt-10 mb-8">
        <Divider orientation="left">
          <Breadcrumb>
            <Breadcrumb.Item> Components </Breadcrumb.Item>
            <Breadcrumb.Item> Colors </Breadcrumb.Item>
          </Breadcrumb>
        </Divider>
      </div>

      {/* Brand Color */}
      <Row className={'mt-6 mb-2'}>
        <Col span={4}>
          <Text type={'title'} level={7} weight={'bold'} >Brand Color</Text>
        </Col>
      </Row>
      <Row className={''}>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--brand-1"><Text type={'title'} level={6} weight={'bold'} extraClass={'pt-4'}>1</Text></div>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--brand-2"><Text type={'title'} level={6} weight={'bold'} extraClass={'pt-4'}>2</Text></div>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--brand-3"><Text type={'title'} level={6} weight={'bold'} extraClass={'pt-4'}>3</Text></div>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--brand-4"><Text type={'title'} level={6} weight={'bold'} extraClass={'pt-4'}>4</Text></div>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--brand-5"><Text type={'title'} level={6} weight={'bold'} extraClass={'pt-4'}>5</Text></div>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--brand-6"><Text type={'title'} level={6} weight={'bold'} extraClass={'pt-4'}>6</Text></div>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--brand-7"><Text type={'title'} level={6} weight={'bold'} extraClass={'pt-4'}>7</Text></div>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--brand-8"><Text type={'title'} level={6} weight={'bold'} extraClass={'pt-4'}>8</Text></div>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--brand-9"><Text type={'title'} level={6} weight={'bold'} extraClass={'pt-4'}>9</Text></div>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--brand-10"><Text type={'title'} level={6} weight={'bold'} extraClass={'pt-4'}>10</Text></div>
      </Row>

      {/* Brand Color */}
      <Row className={'mt-12 mb-2'}>
        <Col span={4}>
          <Text type={'title'} level={7} weight={'bold'} >Neutral/Mono Color Neutral</Text>
        </Col>
      </Row>
      <Row className={''}>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--mono-1"><Text type={'title'} level={6} weight={'bold'} extraClass={'pt-4'}>1</Text></div>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--mono-2"><Text type={'title'} level={6} weight={'bold'} extraClass={'pt-4'}>2</Text></div>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--mono-3"><Text type={'title'} level={6} weight={'bold'} extraClass={'pt-4'}>3</Text></div>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--mono-4"><Text type={'title'} level={6} weight={'bold'} extraClass={'pt-4'}>4</Text></div>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--mono-5"><Text type={'title'} level={6} weight={'bold'} extraClass={'pt-4'}>5</Text></div>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--mono-6"><Text type={'title'} level={6} weight={'bold'} extraClass={'pt-4'}>6</Text></div>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--mono-7"><Text type={'title'} level={6} weight={'bold'} extraClass={'pt-4'}>7</Text></div>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--mono-8"><Text type={'title'} level={6} weight={'bold'} extraClass={'pt-4'}>8</Text></div>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--mono-9"><Text type={'title'} level={6} weight={'bold'} extraClass={'pt-4'}>9</Text></div>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--mono-10"><Text type={'title'} level={6} weight={'bold'} extraClass={'pt-4'}>10</Text></div>
      </Row>

      {/* Functional Colors */}
      <Row className={'mt-12 mb-2'}>
        <Col span={4}>
          <Text type={'title'} level={7} weight={'bold'} >Functional Color</Text>
        </Col>
      </Row>
      <Row className={''}>
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--fn-orange" />
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--fn-red" />
        <div className="mr-4 flex justify-center align-ceter fa-component-color--box fa-component-color--fn-green" />
      </Row>

    </>

  );
}

export default ColorLib;
