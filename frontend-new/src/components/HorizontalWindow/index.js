import { ArrowLeftOutlined, ArrowRightOutlined } from '@ant-design/icons';
import { Button, Carousel } from 'antd';
import React, { useEffect, useState } from 'react';
import styles from './index.module.scss';
import { Text } from 'Components/factorsComponents';
import { FallBackImage } from 'Constants/templates.constants';
export const HorizontalWindowItem = ({ title, image, onClick }) => {
  return (
    <div className={styles.windowItem} onClick={onClick}>
      <img src={image != null ? image : FallBackImage} />
      <div>{title}</div>
    </div>
  );
};
const contentStyle = {
  height: '160px',
  color: '#fff',
  lineHeight: '160px',
  textAlign: 'center',
  background: '#364d79',
  display: 'inherit'
};
const HorizontalWindow = ({ windowTemplates, onWindowClick }) => {
  const windowRef = React.createRef();
  const [allItems, setAllItems] = useState([]);

  useEffect(() => {}, []);
  useEffect(() => {
    let elements = [];
    let length = windowTemplates.length;
    let windowLength = Math.floor(length / 3);
    let odd = length % 3;
    for (let i = 0; i < length - odd; i += 3) {
      elements.push([
        windowTemplates[i],
        windowTemplates[i + 1],
        windowTemplates[i + 2]
      ]);
    }
    let oddarr = [];
    for (let i = 0; i < odd; i++) {
      oddarr.push(windowTemplates[3 * windowLength + i]);
    }
    elements.push(oddarr);
    setAllItems(elements);
  }, [windowTemplates]);
  useEffect(() => {
    console.log(allItems);
  }, [allItems]);
  const onLeftBtn = () => {
    ref.current.prev();
  };
  const onRightBtn = () => {
    ref.current.next();
  };
  let ref = React.createRef();

  return (
    <div className={styles.horizontalWindow}>
      <div className={styles.horizontalWindowTitle}>
        <Text type={'title'} level={6} weight={'bold'} extraClass={`m-0 mr-3`}>
          Related Templates
        </Text>
      </div>
      <div className={styles.horizontalWindowBody}>
        <div className={styles.controls}>
          <Button
            size='large'
            type='text'
            icon={<ArrowLeftOutlined />}
            onClick={onLeftBtn}
          />
        </div>
        <div className={styles.contentWindow} ref={windowRef}>
          <Carousel
            dots={false}
            dotPosition='top'
            style={{ width: '100%', display: 'inherit' }}
            ref={ref}
          >
            {allItems.map((eachWindow, eachIndex) => {
              return (
                <div key={eachIndex} style={contentStyle}>
                  {/* <h1>uhkj</h1> */}
                  {eachWindow.map((eachWindowItem, eachWindowItemIndex) => {
                    return (
                      <HorizontalWindowItem
                        title={eachWindowItem?.title}
                        image={eachWindowItem?.image}
                        onClick={() =>
                          onWindowClick(eachIndex * 3 + eachWindowItemIndex)
                        }
                        key={eachIndex * 3 + eachWindowItemIndex}
                      />
                    );
                  })}
                </div>
              );
            })}
          </Carousel>
          {/* {windowTemplates.map((eachWindow, eachIndex) => {
          return (
            <HorizontalWindowItem
              title={eachWindow.title}
              image={eachWindow.image}
              onClick={() => {
                onWindowClick(eachIndex);
              }}
            />
          );
        })} */}
        </div>
        <div className={styles.controls}>
          <Button
            size='large'
            type='text'
            icon={<ArrowRightOutlined />}
            onClick={onRightBtn}
          />
        </div>
      </div>
    </div>
  );
};

export default HorizontalWindow;
