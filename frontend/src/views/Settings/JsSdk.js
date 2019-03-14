import React, { Component } from 'react';
import { connect } from 'react-redux';
import {
		Col,
		Card,
		CardHeader,
		CardBody,
		Button,
} from 'reactstrap';
import Loading from '../../loading';

const mapStateToProps = store => {
  return {
		projects: store.projects.projects,
		currentProjectId: store.projects.currentProjectId,
  }
}

class JsSdk extends Component {
	constructor(props) {
		super(props);
	}

	getToken() {
    return this.props.projects[this.props.currentProjectId].token;
  }

	getSDKScript() {
    let token = this.getToken();
    let assetURL = BUILD_CONFIG.sdk_asset_url;
    return '(function(c){var s=document.createElement("script");s.type="text/javascript";if(s.readyState){s.onreadystatechange=function(){if(s.readyState=="loaded"||s.readyState=="complete"){s.onreadystatechange=null;c()}}}else{s.onload=function(){c()}}s.src="'+assetURL+'";d=!!document.body?document.body:document.head;d.appendChild(s)})(function(){factors.init("'+token+'")})';
  }

	render() {
		return (
			<Col md='12'>
				<Card className="fapp-bordered-card">
          <CardHeader>
						<button className='btn btn-success' style={{float: 'right', padding: '2px 8px'}}> Copy  <i class='fa fa-copy' style={{marginLeft: '4px', fontWeight: 'inherit'}}></i> </button>
            <strong>Code Snippet</strong>
          </CardHeader>
					<CardBody>
						<div className='fapp-code'>
							<span>{this.getSDKScript()}</span>
						</div>
					</CardBody>
        </Card>
			</Col>
		);
	}
}

export default connect(mapStateToProps, null)(JsSdk);