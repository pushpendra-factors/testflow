import React, { Component } from 'react';

class FunnelArrow extends Component {
  render() {
    var arrowId = "arrow" + this.props.uid;
    return(
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 50 50">
        <defs>
          <marker id={arrowId} markerWidth="10" markerHeight="10" refX="0" refY="3" orient="auto" markerUnits="strokeWidth" viewBox="0 0 20 20">
            <path d="M0,0 L0,6 L9,3 z" fill={this.props.color} />
          </marker>
        </defs>
        <text x="10" y="10">{this.props.conversionString}</text>
        <line x1="0" y1="25" x2="30" y2="25" stroke={this.props.color} strokeWidth="5" markerEnd={"url(#" + arrowId+ ")"} />
      </svg>
    );
  }
}

export default FunnelArrow;
