import React, { FC } from 'react';
import { SVG, Text } from 'Components/factorsComponents';
import { Button, Col, Row } from 'antd';
import styles from './index.module.scss';

interface cardProps {
  bgColor: string;
  title: string;
  description: string;
  learnMoreUrl: string;
  imgUrl: string;
  category: number;
}

function Card({
  bgColor,
  title,
  description,
  learnMoreUrl,
  imgUrl,
  category
}: cardProps) {
  return (
    <div className={`${styles.card}`} style={{ background: `${bgColor}` }}>
      <Row gutter={[24, 24]}>
        <Col span={20}>
          <div className='flex justify-between items-center'>
            <div className='flex flex-col'>
              <Row
                justify='center'
                className={`ml-2 ${category === 1 ? 'mt-6' : 'mt-4'}`}
              >
                <Col span={imgUrl ? 20 : 24}>
                  <Text
                    type='title'
                    level={6}
                    // weight='bold'
                    extraClass='m-0'
                    id='fa-at-text--page-title'
                  >
                    {title}
                  </Text>
                  <Text
                    type='title'
                    level={7}
                    extraClass='m-0 mb-1'
                    color='grey'
                  >
                    {description}
                  </Text>
                  <Button
                    type='link'
                    // size='small'
                    icon={<SVG name='ArrowUpRightSquare' color='blue' />}
                    className={`${styles.learnMoreBtn}`}
                    onClick={() => window.open(learnMoreUrl, '_blank')}
                  >
                    Learn more
                  </Button>
                </Col>
                {imgUrl && (
                  <Col span={4}>
                    <img
                      src={imgUrl}
                      className={`${
                        category === 1 ? styles.catOneImage : styles.image
                      }`}
                      alt=''
                    />
                  </Col>
                )}
              </Row>
            </div>
          </div>
        </Col>
      </Row>
    </div>
  );
}

export default Card;
