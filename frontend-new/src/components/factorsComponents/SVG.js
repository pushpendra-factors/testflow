import React from 'react';
import * as icons from '../svgIcons';

class SVG extends React.Component {
  handleTitleConversion(str) {
    return str?.replace(/(^|\s)\S/g,
      function (t) {
        return t.toUpperCase();
      });
  }

  render() {
    const {
      name, size, width, height, color, extraClass
    } = this.props;
    const properName = this.handleTitleConversion(name) + 'SVG';
    const IconComponent = icons[properName];
    if (!IconComponent) {
      // console.error('Invalid SVG ICON Name --->', name);
      return null;
    }
    const strokeColor =
      color === 'white' ? '#FFFFFF'
        : color === 'black' ? '#0E2647'
          : color === 'purple' ? '#1E89FF' //blue color now.
            : color === 'blue' ? '#1E89FF'
              : color === 'green' ? '#5ACA89'
                : color === 'red' ? '#EA6262'
                  : color === 'grey' ? '#63686F' : color;

    return (
      <IconComponent size={size} width={width} height={height} color={color ? strokeColor : `#63686F` } extraClass={extraClass} />
    );
  }
}

export default SVG;
