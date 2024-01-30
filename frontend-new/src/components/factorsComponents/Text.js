import React from 'react';
import classnames from 'classnames';
import _ from 'lodash';
import { Typography, Tooltip } from 'antd';

const { Title, Paragraph } = Typography;

const textType = {
  title: 'title',
  paragraph: 'paragraph'
};
class Text extends React.Component {
  render() {
    const {
      type,
      level,
      size,
      children,
      weight,
      mini,
      color,
      lineHeight,
      align,
      textCenter,
      isUppercase,
      extraClass,
      truncate = false,
      toolTipTitle,
      charLimit = 30,
      ...otherProps
    } = this.props;

    const defaultFontSize =
      type === textType.paragraph ? (mini ? 7 : 6) : level || size;

    const classList = {
      'fai-text': true,

      // Size
      [`fai-text__size--h${defaultFontSize}`]: true,

      // Weight
      [`fai-text__weight--${weight || 'regular'}`]: true,

      // Color
      [`fai-text__color--${color}`]: color,

      // Line Height
      [`fai-text__line-height--${lineHeight}`]: lineHeight,

      // Alignment
      [`fai-text__align--${align}`]: align,

      // Case
      'fai-text__transform--uppercase': isUppercase,

      [extraClass]: extraClass
    };

    // (Number.isInteger(isSizeDefined)
    // AntD throws error for level>4
    const isSizeDefined = level || size;

    //checks if text truncation and is child is string. ignores if its array.
    const isTextTruncatePossible = truncate && !_.isArray(children);

    const isOverFlow = children?.length > charLimit;

    let truncatedText = children;

    if (isTextTruncatePossible && isOverFlow) {
      truncatedText = `${children.slice(0, charLimit)}${'...'}`;
    }

    if (type === textType.title) {
      const sizeValue = isSizeDefined > 4 ? 4 : isSizeDefined;
      if ((isTextTruncatePossible && isOverFlow) || toolTipTitle) {
        return (
          <Tooltip
            placement={'top'}
            title={toolTipTitle ? toolTipTitle : children}
          >
            <Title
              level={sizeValue}
              {...otherProps}
              className={classnames({ ...classList })}
            >
              {truncatedText}
            </Title>
          </Tooltip>
        );
      } else
        return (
          <Title
            level={sizeValue}
            {...otherProps}
            className={classnames({ ...classList })}
          >
            {children}
          </Title>
        );
    }
    if (type === textType.paragraph) {
      return (
        <Paragraph {...otherProps} className={classnames({ ...classList })}>
          {children}
        </Paragraph>
      );
    } else {
      console.error('Invalid type for Text (Factor-Components)');
      return null;
    }
  }
}

export default Text;
