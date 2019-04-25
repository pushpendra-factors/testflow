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
    return '(function(c){var s=document.createElement("script");s.type="text/javascript";if(s.readyState){s.onreadystatechange=function(){if(s.readyState=="loaded"||s.readyState=="complete"){s.onreadystatechange=null;c()}}}else{s.onload=function(){c()}}s.src="'+assetURL+'";s.async=true;d=!!document.body?document.body:document.head;d.appendChild(s)})(function(){factors.init("'+token+'")})';
	}
	
	renderScriptCode() {
		return (
			<div className='fapp-code'>
				<span style={{display: 'block'}}>{'<script>'}</span>
				<div style={{marginLeft: '15px'}}><span>{this.getSDKScript()}</span></div>
				<span style={{display: 'block'}}>{'</script>'}</span>
			</div>
		);
	}

	render() {
		return (
			<div class='fapp-content fapp-content-margin'>
				<Card className='fapp-bordered-card'>
          <CardHeader>
						{
							/* Todo(Dinesh): Add copy to clipboard, Use the button below. */
							/* <button className='btn btn-success' style={{float: 'right', padding: '2px 8px'}}> Copy  <i class='fa fa-copy' style={{marginLeft: '4px', fontWeight: 'inherit'}}></i> </button> */
						}
            <strong>Code Snippet</strong>
          </CardHeader>
					<CardBody style={{padding: '1.5rem 1.5rem'}}>
						{ this.renderScriptCode() }
					</CardBody>
        </Card>
			</div>
		);
	}
}

export default connect(mapStateToProps, null)(JsSdk);