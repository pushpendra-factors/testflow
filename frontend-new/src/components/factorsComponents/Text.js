import React from 'react';
import classnames from 'classnames';
import { Typography } from 'antd';
const { Title } = Typography;

class Text extends React.Component {
  render() {  
      const {level,size, children, weight, color, lineHeight, align, textCenter, isUppercase, extraClass, ...otherProps} = this.props;  

      const classList = {
        'fai-text': true, 

        //Size
        [`fai-text__size--h${level||size}`]: level||size,
  
        //Weight
        [`fai-text__weight--${weight}`]: weight,
  
        //Color
        [`fai-text__color--${color}`]: color, 
  
        //Line Height
        [`fai-text__line-height--${lineHeight}`]: lineHeight,
  
        //Alignment
        [`fai-text__weight--${align}`]: align,  
  
        //Case
        'fai-text__transform--uppercase': isUppercase,
  
        [extraClass]: extraClass,
      };


    return (
      <>
          <Title level={level||size} {...otherProps} className={classnames({ ...classList })} >{children}</Title>
      </>
    );
  }
}

export default Text;
