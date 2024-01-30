import React, { FC } from 'react';
import { SVG, Text } from 'Components/factorsComponents';
import { Button, Col, Row } from 'antd';
import styles from './index.module.scss';

interface cardProps {
  title: string;
  description: string;
  learnMoreUrl: string;
  imgUrl: string;
}

function Card({ title, description, learnMoreUrl, imgUrl }: cardProps) {
  return (
    <div className={`${styles.card}`}>
      <Row gutter={[24, 24]}>
        <Col span={20}>
          <div className='flex justify-between items-center'>
            <div className='flex flex-col'>
              <Row justify='center' className='pl-6 p-4'>
                <Col span={19}>
                  <Text
                    type='title'
                    level={6}
                    style={{ color: '#000000A6' }}
                    extraClass={`${styles.cardTitle}`}
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
                    icon={
                      <SVG
                        name='ArrowUpRightSquare'
                        extraClass='-mt-1'
                        size={18}
                        color='blue'
                      />
                    }
                    className={`${styles.learnMoreBtn}`}
                    onClick={() => window.open(learnMoreUrl, '_blank')}
                  >
                    Learn more
                  </Button>
                </Col>
                <Col span={5}>
                  <img src={imgUrl} className={`${styles.image}`} alt='' />
                </Col>
              </Row>
            </div>
          </div>
        </Col>
      </Row>
    </div>
  );
}

export default Card;
