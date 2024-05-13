import React from 'react';
import * as icons from '../svgIcons';

class SVG extends React.Component {
  handleTitleConversion(str) {
    return str?.replace(/(^|\s)\S/g, function (t) {
      return t.toUpperCase();
    });
  }

  render() {
    let { name, color } = this.props;
    const properName = this.handleTitleConversion(name) + 'SVG';
    const IconComponent = icons[properName];
    if (!IconComponent) {
      // console.error('Invalid SVG ICON Name --->', name);
      return null;
    }
    // This is added to maintain query Components Icon Colors consistent Globally.
    let queryComponents = {
      events_cq: '#85A5FF',
      funnels_cq: '#FF85C0',
      profiles_cq: '#FFC069',
      KPI_cq: '#FF9C6E',
      Attributions_cq: 'blue'
    };

    if (name in queryComponents) {
      color = queryComponents[name];
    }
    const strokeColor =
      color === 'white'
        ? '#FFFFFF'
        : color === 'black'
          ? '#0E2647'
          : color === 'purple'
            ? '#1E89FF' //blue color now.
            : color === 'blue'
              ? '#1E89FF'
              : color === 'green'
                ? '#5ACA89'
                : color === 'red'
                  ? '#EA6262'
                  : color === 'grey'
                    ? '#63686F'
                    : color;

    return (
      <IconComponent {...this.props} color={color ? strokeColor : `#63686F`} />
    );
  }
}

export default SVG;
