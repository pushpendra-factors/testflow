
import React, { Component } from 'react';
import { connect } from 'react-redux';
import {
		Col,
		Row,		
		Button,
} from 'reactstrap';
import Loading from '../../loading';
import factorsai from '../../common/factorsaiObj';

const mapStateToProps = store => {
  return {
		projects: store.projects.projects,
		currentProjectId: store.projects.currentProjectId,
  }
}

class IosSdk extends Component {
	constructor(props) {
		super(props);
	}

	getToken() {
    return this.props.projects[this.props.currentProjectId].token;
  }

	trackIosSDKRequest = () => {
		factorsai.track('ios_sdk_request', { project_id: this.props.currentProjectId });
	}

	render() {
		return (
			<div className='fapp-content fapp-content-margin'>
				<Row style={{paddingTop: '12%'}}>
						<div style={{width: '100%', textAlign: 'center', fontSize: '32px', fontWeight: '700', color: '#666', marginBottom: '15px', letterSpacing: '0.1rem'}}>Coming Soon<span style={{marginLeft: '3px'}}>!</span></div>
						<div style={{width: '100%', textAlign: 'center'}}>
							<Button style={{ fontSize: '16px', padding: '8px 20px', letterSpacing: '0.05rem' }} color='success' onClick = {this.trackIosSDKRequest} >Request</Button>
						</div>
				</Row>
			</div>
		);
	}
}

export default connect(mapStateToProps, null)(IosSdk);