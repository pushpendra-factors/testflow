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

	render() {
		return (
			<div class='fapp-content fapp-content-margin'>
				<Card className='fapp-bordered-card'>
          <CardHeader>						
            <strong>Android .aar</strong>
          </CardHeader>
					<CardBody style={{padding: '1.5rem 1.5rem'}}>
						<p><b>Project Token:</b> { this.getToken() }</p>
						<p><a href={BUILD_CONFIG.android_sdk_asset_url} download>Download</a></p>
						<div className='fapp-code'>						
							<div style={{marginLeft: '15px'}}>
								<span>
								  // Add permissions in Android Manifest file<br></br>
									android.permission.INTERNET <br></br>
									android.permission.ACCESS_NETWORK_STATE <br></br><br></br>
									// Add Dependency in build.gradle(app) file<br></br>
									implementation 'com.squareup.okhttp3:okhttp:3.9.0'<br></br><br></br>
									Copy .aar file in libs folder<br></br>
									Import the .aar file (https://stackoverflow.com/a/34919810)<br></br><br></br>
									// Init FactorsClient <br></br>
									FactorsClient.getInstance().init(this, "{this.getToken()}");<br></br><br></br>
									// Optional Set log level, Log.INFO is default<br></br>
									FactorsClient.getInstance().setLogLevel(Log.INFO);<br></br><br></br>									
									// Track Events<br></br>
									FactorsClient.getInstance().track("click up", new JSONObject().put("key","up"));<br></br>
								</span>	
							</div>
						</div>
					</CardBody>
        </Card>
			</div>
		);
	}
}

export default connect(mapStateToProps, null)(AndroidSdk);