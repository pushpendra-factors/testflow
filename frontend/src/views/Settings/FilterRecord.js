import React, { Component } from 'react';
import {
    Row,
    Col,
    Input,
    Button,
} from 'reactstrap';


class FilterRecord extends Component {
  constructor(props) {
    super(props);
  }

  render() {
    return (
      <Row style={{padding: "10px 0"}}>
        <Col md={{size: 4}}>
          <Input type="text" value={this.props.domain} className="fapp-input-disabled" readOnly />
        </Col>
        <Col md={{size: 4}}>
          <Input type="text" value={this.props.expr} className="fapp-input-disabled" readOnly />
        </Col>
        <Col md={{size: 3}}>
        < Input type="text" value={this.props.name} onChange={this.props.handleEventNameChange}/>
        </Col>
        <Col>
          <Button className="fapp-inline-button" ><i className="icon-check" onClick={this.props.handleUpdate} style={{color: this.props.getUpdateButtonColor()}}></i></Button>
          <Button className="fapp-inline-button"><i className="icon-trash" onClick={this.props.handleDelete}></i></Button>
        </Col>
      </Row>
    )
  }
}

export default FilterRecord;