/* eslint-disable */
import React from 'react';
import {
  Layout, Breadcrumb, Row, Col, Divider
} from 'antd';
import { Text } from '../../components/factorsComponents';

function TextLib() {
  const { Content } = Layout;
  return (
    <>
      <div className="mt-20 mb-8">
        <Divider orientation="left">
          <Breadcrumb>
            <Breadcrumb.Item> Components </Breadcrumb.Item>
            <Breadcrumb.Item> Text </Breadcrumb.Item>
          </Breadcrumb>
        </Divider>
      </div>

      <Row>
        <Col span={18}>

          <Text type={'title'} level={1} weight={'bold'}>Heading Style - Title1/40</Text>
          <Text type={'title'} level={2} weight={'bold'}>Heading Style - Title2/32</Text>
          <Text type={'title'} level={3} weight={'bold'}>Heading Style - Title3/24</Text>
          <Text type={'title'} level={4} >Heading Style - Title4/20</Text>
          <Text type={'title'} level={5} >Heading Style - Title5/18</Text>
          <Text type={'title'} level={6} >Heading Style - Title6/16</Text>

          <Text type={'title'} level={6} weight={'bold'} extraClass={'mt-8'} >Use for Headings - subtitle1/16</Text>
          <Text type={'title'} level={7} weight={'bold'} extraClass={'my-2'}>Use for Headings - subtitle2/14</Text>

          <Text type={'paragraph'} extraClass={'mt-8'}>Leverage agile frameworks to provide a robust synopsis for high level overviews. Iterative approaches to corporate strategy foster collaborative thinking to further the overall value proposition. Organically grow the holistic world view of disruptive innovation via workplace diversity and empowerment.</Text>

          <Text type={'paragraph'} ellipsis extraClass={'my-4'}>Leverage agile frameworks to provide a robust synopsis for high level overviews. Iterative approaches to corporate strategy foster collaborative thinking to further the overall value proposition. Organically grow the holistic world view of disruptive innovation via workplace diversity and empowerment.</Text>
          <Text type={'paragraph'} mini extraClass={'my-4'}>Leverage agile frameworks to provide a robust synopsis for high level overviews. Iterative approaches to corporate strategy foster collaborative thinking to further the overall value proposition. Organically grow the holistic world view of disruptive innovation via workplace diversity and empowerment.</Text>

        </Col>
      </Row>

    </>

  );
}

export default TextLib;
