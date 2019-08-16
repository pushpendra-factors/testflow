import React, { Component } from 'react';
import onClickOutside from 'react-onclickoutside';
import { DateRangePicker } from 'react-date-range'; 

class DateRangePickerWithCloseHandler extends Component { 
  constructor(props) {
    super(props);
  }

  handleClickOutside = () => {
    this.props.closeDatePicker();
  }

  render() {
    return <DateRangePicker {...this.props} />
  }
}
  
export default onClickOutside(DateRangePickerWithCloseHandler);