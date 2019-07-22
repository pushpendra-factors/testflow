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

class AndroidSdk extends Component {
	constructor(props) {
		super(props);
	}

	getToken() {
    	return this.props.projects[this.props.currentProjectId].token;
	}

	renderCard() {
		return (
			<Card className='fapp-bordered-card'>
				<CardHeader>						
					<strong>Android SDK</strong>
				</CardHeader>
				<CardBody style={{padding: '1.5rem 1.5rem'}}>
					<p><a href={BUILD_CONFIG.android_sdk_asset_url} download>Download file (.aar)</a></p>
					<p className='card-text'>Add the below code to your app by following instructions given as comments.</p>
					<div className='fapp-code'>				
						<p className='green'>// Add permissions in Android Manifest file</p>
						<p className='blue'>android.permission.INTERNET</p>
						<p className='blue'>android.permission.ACCESS_NETWORK_STATE</p>
						<p className='green'>// Add Dependency in build.gradle(app) file</p>
						<p className='blue'>implementation 'com.squareup.okhttp3:okhttp:3.9.0'</p>
						<p className='green'> // Copy .aar file in libs folder</p>
						<p className='green'> // Import the .aar file (https://stackoverflow.com/a/34919810)</p>
						<p className='green'>// Init FactorsClient </p>
						<p className='blue'>FactorsClient.getInstance().init(this, "<span className='red'>{this.getToken()}</span>");</p>
						<p className='green'>// Optional Set log level, Log.INFO is default</p>
						<p className='blue'>FactorsClient.getInstance().setLogLevel(Log.INFO);</p>								
						<p className='green'>// Track Events</p>
						<p className='blue'>FactorsClient.getInstance().track("click up", new JSONObject().put("key","up"));</p>
					</div>
				</CardBody>
			</Card>
		);
	}

	render() {
		console.log(this.props.cardOnly);
		if (this.props.cardOnly) return this.renderCard();

		return (
			<div class='fapp-content fapp-content-margin'>
				{ this.renderCard() }
			</div>
		);
	}
}

export default connect(mapStateToProps, null)(AndroidSdk);