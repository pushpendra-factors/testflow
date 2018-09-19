import React, { Component } from 'react';

class FunnelArrow extends Component {
  render() {
    var arrowId = "arrow" + this.props.uid;
    return(
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
        <defs>
          <marker id={arrowId} markerWidth="10" markerHeight="10" refX="0" refY="3" orient="auto" markerUnits="strokeWidth" viewBox="0 0 20 20">
            <path d="M0,0 L0,6 L9,3 z" fill={this.props.color} />
          </marker>
        </defs>
        <line x1="10" y1="50" x2="70" y2="50" stroke={this.props.color} strokeWidth="5" markerEnd={"url(#" + arrowId+ ")"} />
      </svg>
    );
  }
}

export default FunnelArrow;
