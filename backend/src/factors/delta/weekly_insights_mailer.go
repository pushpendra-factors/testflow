package delta

import (
	C "factors/config"
	"factors/model/store"
	"fmt"
	"time"
	U "factors/util"
	log "github.com/sirupsen/logrus"
)

func MailWeeklyInsights(projectID int64, configs map[string]interface{}) (map[string]interface{}, bool) {
	result := make(map[string]interface{})
	if configs["endTimestamp"].(int64) > U.TimeNowUnix() {
		result["error"] = "invalid end timestamp"
		return result, false
	}
	wetRun := configs["wetRun"].(bool)
	dashboards := StaticDashboardListForMailer()
	startTimestamp := configs["startTimestamp"].(int64)
	projectAgents, _ := store.GetStore().GetProjectAgentMappingsByProjectId(projectID)
	for _, agent := range projectAgents {
		sub := "Weekly Insights Digest"
		agent, _ := store.GetStore().GetAgentByUUID(agent.AgentUUID)
		agentName := agent.FirstName
		for _, dashboard := range dashboards {
			queryId := dashboard.QueryId
			startTimestampTimeFormat := time.Unix(startTimestamp, 0)
			startTimestampMinusOneweekTimeFormat := time.Unix(startTimestamp-(86400*7), 0)
			responseResult, _ := GetWeeklyInsights(projectID, "", queryId, &startTimestampTimeFormat, &startTimestampMinusOneweekTimeFormat, "w", 11, 100, 1, true)
			response := (responseResult).(WeeklyInsights)
			digestHeader := GetQueryHeader(queryId)
			headerValue1 := response.Goal.W1
			headerValue2 := response.Goal.W2
			headerPerc := fmt.Sprintf("%.2f%s", response.Goal.Percentage ,"%%")
			line1Value1 := response.Insights[0].Key
			line1Value2 := response.Insights[0].Value
			line1Perc := fmt.Sprintf("%.2f%s", response.Insights[0].ActualValues.Percentage ,"%%")
			line2Value1 := response.Insights[1].Key
			line2Value2 := response.Insights[1].Value
			line2Perc := fmt.Sprintf("%.2f%s", response.Insights[1].ActualValues.Percentage ,"%%")
			line3Value1 := response.Insights[2].Key
			line3Value2 := response.Insights[2].Value
			line3Perc := fmt.Sprintf("%.2f%s", response.Insights[2].ActualValues.Percentage ,"%%")
			html := returnWIMailerTemplate(agentName, digestHeader, headerPerc, headerValue1, headerValue2, line1Value1, line1Value2, line1Perc, line2Value1, line2Value2, line2Perc, line3Value1, line3Value2, line3Perc)
			
			if(wetRun == true){
				err := C.GetServices().Mailer.SendMail(agent.Email, C.GetFactorsSenderEmail(), sub, html, "")
				if err != nil {
					log.WithError(err).Error("failed to send email alert")
					continue
				}
			} else {
				log.Info(html)
				log.Info(agentName, digestHeader, headerPerc, headerValue1, headerValue2, line1Value1, line1Value2, line1Perc, line2Value1, line2Value2, line2Perc, line3Value1, line3Value2, line3Perc)
			}
		}
		
	}
	return result, true
}

func returnWIMailerTemplate(agentName string, digestHeader string, headerPerc string, headerValue1 float64, headerValue2 float64, line1Value1 string, line1Value2 string, line1Perc string, line2Value1 string, line2Value2 string, line2Perc string, line3Value1 string, line3Value2 string, line3Perc string)(string){
	template := fmt.Sprintf(`<<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
	<html
	  xmlns="http://www.w3.org/1999/xhtml"
	  lang="en"
	  xml:lang="en"
	  style="background: #efefef !important; font-family: 'Open Sans', sans-serif"
	>
	  <head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
		<meta name="viewport" content="width=device-width" />
		<title>[object Object]</title>
		<style>
		  @media only screen {
			html {
			  min-height: 100%%;
			  background: #b7bec8;
			}
		  }
		  @media only screen and (max-width: 596px) {
			table.body img {
			  width: auto;
			  height: auto;
			}
			table.body center {
			  min-width: 0 !important;
			}
			table.body .container {
			  width: 95%% !important;
			}
			table.body .columns {
			  height: auto !important;
			  -moz-box-sizing: border-box;
			  -webkit-box-sizing: border-box;
			  box-sizing: border-box;
			  padding-left: 16px !important;
			  padding-right: 16px !important;
			}
			table.body .columns .columns {
			  padding-left: 0 !important;
			  padding-right: 0 !important;
			}
			th.small-12 {
			  display: inline-block !important;
			  width: 100%% !important;
			}
			.columns th.small-12 {
			  display: block !important;
			  width: 100%% !important;
			}
			table.menu {
			  width: 100%% !important;
			}
			table.menu td,
			table.menu th {
			  width: auto !important;
			  display: inline-block !important;
			}
			table.menu.vertical td,
			table.menu.vertical th {
			  display: block !important;
			}
			table.menu[align="center"] {
			  width: auto !important;
			}
		  }
		</style>
	  </head>
	  <body
		style="
		  -moz-box-sizing: border-box;
		  -ms-text-size-adjust: 100%%;
		  -webkit-box-sizing: border-box;
		  -webkit-text-size-adjust: 100%%;
		  margin: 0;
		  background: #efefef !important;
		  box-sizing: border-box;
		  color: #3e516c;
		  font-family: 'Open Sans', sans-serif;
		  font-size: 16px;
		  font-weight: 400;
		  line-height: 1.5;
		  list-style: 1.5;
		  margin: 0;
		  min-width: 100%%;
		  padding: 0;
		  padding-bottom: 0;
		  padding-left: 0;
		  padding-right: 0;
		  padding-top: 0;
		  text-align: left;
		  width: 100%% !important;
		"
	  >
		<span
		  class="preheader"
		  style="
			color: #b7bec8;
			display: none !important;
			font-size: 1px;
			line-height: 1px;
			max-height: 0;
			max-width: 0;
			mso-hide: all !important;
			opacity: 0;
			overflow: hidden;
			visibility: hidden;
		  "
		></span>
		<table
		  class="body"
		  style="
			margin: 0;
			background: #efefef !important;
			border-collapse: collapse;
			border-spacing: 0;
			color: #3e516c;
			font-family: 'Open Sans', sans-serif;
			font-size: 16px;
			font-weight: 400;
			height: 100%%;
			line-height: 1.5;
			margin: 0;
			padding-bottom: 0;
			padding-left: 0;
			padding-right: 0;
			padding-top: 0;
			text-align: left;
			vertical-align: top;
			width: 100%%;
		  "
		>
		  <tr
			style="
			  padding-bottom: 0;
			  padding-left: 0;
			  padding-right: 0;
			  padding-top: 0;
			  text-align: left;
			  vertical-align: top;
			"
		  >
			<td
			  class="center"
			  align="center"
			  valign="top"
			  style="
				-moz-hyphens: auto;
				-webkit-hyphens: auto;
				margin: 0;
				border-collapse: collapse !important;
				color: #3e516c;
				font-family: 'Open Sans', sans-serif;
				font-size: 16px;
				font-weight: 400;
				hyphens: auto;
				line-height: 1.5;
				margin: 0;
				padding-bottom: 0;
				padding-left: 0;
				padding-right: 0;
				padding-top: 0;
				text-align: left;
				vertical-align: top;
				word-wrap: break-word;
			  "
			>
			  <center style="min-width: 580px; width: 100%%">
				<table
				  align="center"
				  class="spacer float-center"
				  style="
					margin: 0 auto;
					border-collapse: collapse;
					border-spacing: 0;
					float: none;
					margin: 0 auto;
					padding-bottom: 0;
					padding-left: 0;
					padding-right: 0;
					padding-top: 0;
					text-align: center;
					vertical-align: top;
					width: 100%%;
				  "
				>
				  <tbody>
					<tr
					  style="
						padding-bottom: 0;
						padding-left: 0;
						padding-right: 0;
						padding-top: 0;
						text-align: left;
						vertical-align: top;
					  "
					>
					  <td
						height="16"
						style="
						  -moz-hyphens: auto;
						  -webkit-hyphens: auto;
						  margin: 0;
						  border-collapse: collapse !important;
						  color: #3e516c;
						  font-family: 'Open Sans', sans-serif;
						  font-size: 16px;
						  font-weight: 400;
						  hyphens: auto;
						  line-height: 16px;
						  margin: 0;
						  mso-line-height-rule: exactly;
						  padding-bottom: 0;
						  padding-left: 0;
						  padding-right: 0;
						  padding-top: 0;
						  text-align: left;
						  vertical-align: top;
						  word-wrap: break-word;
						"
					  >
						&nbsp;
					  </td>
					</tr>
				  </tbody>
				</table>
				<table
				  align="center"
				  class="container body-border float-center"
				  style="
					margin: 0 auto;
					background: #fefefe;
					border-collapse: collapse;
					border-spacing: 0;
					float: none;
					margin: 0 auto;
					padding-bottom: 0;
					padding-left: 0;
					padding-right: 0;
					padding-top: 0;
					text-align: center;
					vertical-align: top;
					width: 580px;
				  "
				>
				  <tbody>
					<tr
					  style="
						padding-bottom: 0;
						padding-left: 0;
						padding-right: 0;
						padding-top: 0;
						text-align: left;
						vertical-align: top;
					  "
					>
					  <td
						style="
						  -moz-hyphens: auto;
						  -webkit-hyphens: auto;
						  margin: 0;
						  border-collapse: collapse !important;
						  color: #3e516c;
						  font-family: 'Open Sans', sans-serif;
						  font-size: 16px;
						  font-weight: 400;
						  hyphens: auto;
						  line-height: 1.5;
						  margin: 0;
						  padding-bottom: 0;
						  padding-left: 0;
						  padding-right: 0;
						  padding-top: 0;
						  text-align: left;
						  vertical-align: top;
						  word-wrap: break-word;
						"
					  >
						<table
						  class="row"
						  style="
							border-collapse: collapse;
							border-spacing: 0;
							display: table;
							padding: 0;
							padding-bottom: 0;
							padding-left: 0;
							padding-right: 0;
							padding-top: 0;
							position: relative;
							text-align: left;
							vertical-align: top;
							width: 100%%;
						  "
						>
						  <tbody>
							<tr
							  style="
								padding-bottom: 0;
								padding-left: 0;
								padding-right: 0;
								padding-top: 0;
								text-align: left;
								vertical-align: top;
							  "
							>
							  <th
								class="small-12 large-2 columns first"
								style="
								  -moz-hyphens: auto;
								  -webkit-hyphens: auto;
								  margin: 0 auto;
								  border-collapse: collapse !important;
								  color: #3e516c;
								  font-family: 'Open Sans', sans-serif;
								  font-size: 16px;
								  font-weight: 400;
								  hyphens: auto;
								  line-height: 1.5;
								  margin: 0 auto;
								  padding-bottom: 16px;
								  padding-left: 16px;
								  padding-right: 8px;
								  padding-top: 0;
								  text-align: left;
								  vertical-align: top;
								  width: 80.67px;
								  word-wrap: break-word;
								"
							  >
								<table
								  style="
									border-collapse: collapse;
									border-spacing: 0;
									padding-bottom: 0;
									padding-left: 0;
									padding-right: 0;
									padding-top: 0;
									text-align: left;
									vertical-align: top;
									width: 100%%;
								  "
								>
								  <tbody>
									<tr
									  style="
										padding-bottom: 0;
										padding-left: 0;
										padding-right: 0;
										padding-top: 0;
										text-align: left;
										vertical-align: top;
									  "
									>
									  <th
										style="
										  -moz-hyphens: auto;
										  -webkit-hyphens: auto;
										  margin: 0;
										  border-collapse: collapse !important;
										  color: #3e516c;
										  font-family: 'Open Sans', sans-serif;
										  font-size: 16px;
										  font-weight: 400;
										  hyphens: auto;
										  line-height: 1.5;
										  margin: 0;
										  padding-bottom: 0;
										  padding-left: 0;
										  padding-right: 0;
										  padding-top: 0;
										  text-align: left;
										  vertical-align: top;
										  word-wrap: break-word;
										"
									  ></th>
									</tr>
								  </tbody>
								</table>
							  </th>
							  <th
								class="small-12 large-8 columns"
								style="
								  -moz-hyphens: auto;
								  -webkit-hyphens: auto;
								  margin: 0 auto;
								  border-collapse: collapse !important;
								  color: #3e516c;
								  font-family: 'Open Sans', sans-serif;
								  font-size: 16px;
								  font-weight: 400;
								  hyphens: auto;
								  line-height: 1.5;
								  margin: 0 auto;
								  padding-bottom: 16px;
								  padding-left: 8px;
								  padding-right: 8px;
								  padding-top: 0;
								  text-align: left;
								  vertical-align: top;
								  width: 370.67px;
								  word-wrap: break-word;
								"
							  >
								<table
								  style="
									border-collapse: collapse;
									border-spacing: 0;
									padding-bottom: 0;
									padding-left: 0;
									padding-right: 0;
									padding-top: 0;
									text-align: left;
									vertical-align: top;
									width: 100%%;
								  "
								>
								  <tbody>
									<tr
									  style="
										padding-bottom: 0;
										padding-left: 0;
										padding-right: 0;
										padding-top: 0;
										text-align: left;
										vertical-align: top;
									  "
									>
									  <th
										style="
										  -moz-hyphens: auto;
										  -webkit-hyphens: auto;
										  margin: 0;
										  border-collapse: collapse !important;
										  color: #3e516c;
										  font-family: 'Open Sans', sans-serif;
										  font-size: 16px;
										  font-weight: 400;
										  hyphens: auto;
										  line-height: 1.5;
										  margin: 0;
										  padding-bottom: 0;
										  padding-left: 0;
										  padding-right: 0;
										  padding-top: 0;
										  text-align: left;
										  vertical-align: top;
										  word-wrap: break-word;
										"
									  >
										<table
										  class="spacer"
										  style="
											border-collapse: collapse;
											border-spacing: 0;
											padding-bottom: 0;
											padding-left: 0;
											padding-right: 0;
											padding-top: 0;
											text-align: left;
											vertical-align: top;
											width: 100%%;
										  "
										>
										  <tbody>
											<tr
											  style="
												padding-bottom: 0;
												padding-left: 0;
												padding-right: 0;
												padding-top: 0;
												text-align: left;
												vertical-align: top;
											  "
											>
											  <td
												height="48"
												style="
												  -moz-hyphens: auto;
												  -webkit-hyphens: auto;
												  margin: 0;
												  border-collapse: collapse !important;
												  color: #3e516c;
												  font-family: 'Open Sans',
													sans-serif;
												  font-size: 48px;
												  font-weight: 400;
												  hyphens: auto;
												  line-height: 48px;
												  margin: 0;
												  mso-line-height-rule: exactly;
												  padding-bottom: 0;
												  padding-left: 0;
												  padding-right: 0;
												  padding-top: 0;
												  text-align: left;
												  vertical-align: top;
												  word-wrap: break-word;
												"
											  >
												&nbsp;
											  </td>
											</tr>
										  </tbody>
										</table>
										<center
										  style="min-width: 338.67px; width: 100%%"
										>
										  <img
											class="brand-logo float-center"
											src="https://s3.amazonaws.com/www.factors.ai/email-templates/images/factors-logo.png"
											align="center"
											style="
											  -ms-interpolation-mode: bicubic;
											  margin: 0 auto;
											  clear: both;
											  display: block;
											  float: none;
											  margin: 0 auto;
											  max-width: 175px;
											  outline: 0;
											  text-align: center;
											  text-decoration: none;
											  width: auto;
											"
										  />
										</center>
										<table
										  class="spacer"
										  style="
											border-collapse: collapse;
											border-spacing: 0;
											padding-bottom: 0;
											padding-left: 0;
											padding-right: 0;
											padding-top: 0;
											text-align: left;
											vertical-align: top;
											width: 100%%;
										  "
										>
										  <tbody>
											<tr
											  style="
												padding-bottom: 0;
												padding-left: 0;
												padding-right: 0;
												padding-top: 0;
												text-align: left;
												vertical-align: top;
											  "
											>
											  <td
												height="48"
												style="
												  -moz-hyphens: auto;
												  -webkit-hyphens: auto;
												  margin: 0;
												  border-collapse: collapse !important;
												  color: #3e516c;
												  font-family: 'Open Sans',
													sans-serif;
												  font-size: 48px;
												  font-weight: 400;
												  hyphens: auto;
												  line-height: 48px;
												  margin: 0;
												  mso-line-height-rule: exactly;
												  padding-bottom: 0;
												  padding-left: 0;
												  padding-right: 0;
												  padding-top: 0;
												  text-align: left;
												  vertical-align: top;
												  word-wrap: break-word;
												"
											  >
												&nbsp;
											  </td>
											</tr>
										  </tbody>
										</table>
										<center
										  style="min-width: 338.67px; width: 100%%"
										>
										  <p
											class="text-center h1 float-center"
											align="center"
											style="
											  margin: 0;
											  margin-bottom: 10px;
											  color: #3e516c;
											  font-family: 'Open Sans', sans-serif;
											  font-size: 16px;
											  font-weight: 400;
											  line-height: 1.5;
											  margin: 0;
											  margin-bottom: 10px;
											  padding-bottom: 0;
											  padding-left: 0;
											  padding-right: 0;
											  padding-top: 0;
											  text-align: center;
											"
										  >
											<span
											  class="text-bold"
											  style="font-weight: 700"
											  >Hey %v </span
											>, here’s the latest edition of your
											<span
											  class="text-bold"
											  style="font-weight: 700"
											  >Factors digest</span
											>, covering the performance of your top
											goals.
										  </p>
										  <table
											align="center"
											class="spacer float-center"
											style="
											  margin: 0 auto;
											  border-collapse: collapse;
											  border-spacing: 0;
											  float: none;
											  margin: 0 auto;
											  padding-bottom: 0;
											  padding-left: 0;
											  padding-right: 0;
											  padding-top: 0;
											  text-align: center;
											  vertical-align: top;
											  width: 100%%;
											"
										  >
											<tbody>
											  <tr
												style="
												  padding-bottom: 0;
												  padding-left: 0;
												  padding-right: 0;
												  padding-top: 0;
												  text-align: left;
												  vertical-align: top;
												"
											  >
												<td
												  height="25"
												  style="
													-moz-hyphens: auto;
													-webkit-hyphens: auto;
													margin: 0;
													border-collapse: collapse !important;
													color: #3e516c;
													font-family: 'Open Sans',
													  sans-serif;
													font-size: 25px;
													font-weight: 400;
													hyphens: auto;
													line-height: 25px;
													margin: 0;
													mso-line-height-rule: exactly;
													padding-bottom: 0;
													padding-left: 0;
													padding-right: 0;
													padding-top: 0;
													text-align: left;
													vertical-align: top;
													word-wrap: break-word;
												  "
												>
												  &nbsp;
												</td>
											  </tr>
											</tbody>
										  </table>
										  <table
											align="center"
											class="row border float-center"
											style="
											  margin: 0 auto;
											  border: 1px solid #e7e9ed;
											  border-collapse: collapse;
											  border-radius: 5px !important;
											  border-spacing: 0;
											  display: table;
											  float: none;
											  margin: 0 auto;
											  padding: 0;
											  padding-bottom: 0;
											  padding-left: 0;
											  padding-right: 0;
											  padding-top: 0;
											  position: relative;
											  text-align: center;
											  vertical-align: top;
											  width: 100%%;
											"
										  >
											<tbody>
											  <tr
												style="
												  padding-bottom: 0;
												  padding-left: 0;
												  padding-right: 0;
												  padding-top: 0;
												  text-align: left;
												  vertical-align: top;
												"
											  >
												<th
												  class="small-12 large-12 columns first last"
												  style="
													-moz-hyphens: auto;
													-webkit-hyphens: auto;
													margin: 0 auto;
													border-collapse: collapse !important;
													color: #3e516c;
													font-family: 'Open Sans',
													  sans-serif;
													font-size: 16px;
													font-weight: 400;
													hyphens: auto;
													line-height: 1.5;
													margin: 0 auto;
													padding-bottom: 16px;
													padding-left: 0 !important;
													padding-right: 0 !important;
													padding-top: 0;
													text-align: left;
													vertical-align: top;
													width: 100%%;
													word-wrap: break-word;
												  "
												>
												  <table
													style="
													  border-collapse: collapse;
													  border-spacing: 0;
													  padding-bottom: 0;
													  padding-left: 0;
													  padding-right: 0;
													  padding-top: 0;
													  text-align: left;
													  vertical-align: top;
													  width: 100%%;
													"
												  >
													<tbody>
													  <tr
														style="
														  padding-bottom: 0;
														  padding-left: 0;
														  padding-right: 0;
														  padding-top: 0;
														  text-align: left;
														  vertical-align: top;
														"
													  >
														<th
														  style="
															-moz-hyphens: auto;
															-webkit-hyphens: auto;
															margin: 0;
															border-collapse: collapse !important;
															color: #3e516c;
															font-family: 'Open Sans',
															  sans-serif;
															font-size: 16px;
															font-weight: 400;
															hyphens: auto;
															line-height: 1.5;
															margin: 0;
															padding-bottom: 0;
															padding-left: 0;
															padding-right: 0;
															padding-top: 0;
															text-align: left;
															vertical-align: top;
															word-wrap: break-word;
														  "
														>
														  <table
															class="row"
															style="
															  border-collapse: collapse;
															  border-spacing: 0;
															  display: table;
															  padding: 0;
															  padding-bottom: 0;
															  padding-left: 0;
															  padding-right: 0;
															  padding-top: 0;
															  position: relative;
															  text-align: left;
															  vertical-align: top;
															  width: 100%%;
															"
														  >
															<tbody>
															  <tr
																style="
																  padding-bottom: 0;
																  padding-left: 0;
																  padding-right: 0;
																  padding-top: 0;
																  text-align: left;
																  vertical-align: top;
																"
															  >
																<th
																  class="small-12 large-6 columns first"
																  style="
																	-moz-hyphens: auto;
																	-webkit-hyphens: auto;
																	margin: 0 auto;
																	border-collapse: collapse !important;
																	color: #3e516c;
																	font-family: 'Open Sans',
																	  sans-serif;
																	font-size: 16px;
																	font-weight: 400;
																	hyphens: auto;
																	line-height: 1.5;
																	margin: 0 auto;
																	padding-bottom: 16px;
																	padding-left: 0 !important;
																	padding-right: 8px;
																	padding-top: 0;
																	text-align: left;
																	vertical-align: top;
																	width: 50%%;
																	word-wrap: break-word;
																  "
																>
																  <table
																	style="
																	  border-collapse: collapse;
																	  border-spacing: 0;
																	  padding-bottom: 0;
																	  padding-left: 0;
																	  padding-right: 0;
																	  padding-top: 0;
																	  text-align: left;
																	  vertical-align: top;
																	  width: 100%%;
																	"
																  >
																	<tbody>
																	  <tr
																		style="
																		  padding-bottom: 0;
																		  padding-left: 0;
																		  padding-right: 0;
																		  padding-top: 0;
																		  text-align: left;
																		  vertical-align: top;
																		"
																	  >
																		<th
																		  style="
																			-moz-hyphens: auto;
																			-webkit-hyphens: auto;
																			margin: 0;
																			border-collapse: collapse !important;
																			color: #3e516c;
																			font-family: 'Open Sans',
																			  sans-serif;
																			font-size: 16px;
																			font-weight: 400;
																			hyphens: auto;
																			line-height: 1.5;
																			margin: 0;
																			padding-bottom: 0;
																			padding-left: 0;
																			padding-right: 0;
																			padding-top: 0;
																			text-align: left;
																			vertical-align: top;
																			word-wrap: break-word;
																		  "
																		>
																		  <p
																			class="text-left margin-left margin-top text-bold h4"
																			style="
																			  margin: 0;
																			  margin-bottom: 10px;
																			  color: #3e516c;
																			  font-family: 'Open Sans',
																				sans-serif;
																			  font-size: 16px;
																			  font-weight: 700;
																			  line-height: 1.5;
																			  margin: 0;
																			  margin-bottom: 10px;
																			  margin-left: 15px;
																			  margin-top: 20px;
																			  padding-bottom: 0;
																			  padding-left: 0;
																			  padding-right: 0;
																			  padding-top: 0;
																			  text-align: left;
																			"
																		  >
																			%v
																		  </p>
																		  <p
																			class="text-left margin-left h7 medium-grey"
																			style="
																			  margin: 0;
																			  margin-bottom: 10px;
																			  color: #8692a3;
																			  font-family: 'Open Sans',
																				sans-serif;
																			  font-size: 14px;
																			  font-weight: 400;
																			  line-height: 1.5;
																			  margin: 0;
																			  margin-bottom: 10px;
																			  margin-left: 15px;
																			  padding-bottom: 0;
																			  padding-left: 0;
																			  padding-right: 0;
																			  padding-top: 0;
																			  text-align: left;
																			"
																		  >
																			Since
																			Last
																			week
																		  </p>
																		</th>
																	  </tr>
																	</tbody>
																  </table>
																</th>
																<th
																  class="small-12 large-6 columns last"
																  style="
																	-moz-hyphens: auto;
																	-webkit-hyphens: auto;
																	margin: 0 auto;
																	border-collapse: collapse !important;
																	color: #3e516c;
																	font-family: 'Open Sans',
																	  sans-serif;
																	font-size: 16px;
																	font-weight: 400;
																	hyphens: auto;
																	line-height: 1.5;
																	margin: 0 auto;
																	padding-bottom: 16px;
																	padding-left: 8px;
																	padding-right: 0 !important;
																	padding-top: 0;
																	text-align: left;
																	vertical-align: top;
																	width: 50%%;
																	word-wrap: break-word;
																  "
																>
																  <table
																	style="
																	  border-collapse: collapse;
																	  border-spacing: 0;
																	  padding-bottom: 0;
																	  padding-left: 0;
																	  padding-right: 0;
																	  padding-top: 0;
																	  text-align: left;
																	  vertical-align: top;
																	  width: 100%%;
																	"
																  >
																	<tbody>
																	  <tr
																		style="
																		  padding-bottom: 0;
																		  padding-left: 0;
																		  padding-right: 0;
																		  padding-top: 0;
																		  text-align: left;
																		  vertical-align: top;
																		"
																	  >
																		<th
																		  style="
																			-moz-hyphens: auto;
																			-webkit-hyphens: auto;
																			margin: 0;
																			border-collapse: collapse !important;
																			color: #3e516c;
																			font-family: 'Open Sans',
																			  sans-serif;
																			font-size: 16px;
																			font-weight: 400;
																			hyphens: auto;
																			line-height: 1.5;
																			margin: 0;
																			padding-bottom: 0;
																			padding-left: 0;
																			padding-right: 0;
																			padding-top: 0;
																			text-align: left;
																			vertical-align: top;
																			word-wrap: break-word;
																		  "
																		>
																		  <p
																			class="text-right margin-right margin-top text-bold h4"
																			style="
																			  margin: 0;
																			  margin-bottom: 10px;
																			  color: #3e516c;
																			  font-family: 'Open Sans',
																				sans-serif;
																			  font-size: 16px;
																			  font-weight: 700;
																			  line-height: 1.5;
																			  margin: 0;
																			  margin-bottom: 10px;
																			  margin-right: 15px;
																			  margin-top: 20px;
																			  padding-bottom: 0;
																			  padding-left: 0;
																			  padding-right: 0;
																			  padding-top: 0;
																			  text-align: right;
																			"
																		  >
																			%v
																		  </p>
																		  <p
																			class="text-right margin-right h7 medium-grey"
																			style="
																			  margin: 0;
																			  margin-bottom: 10px;
																			  color: #8692a3;
																			  font-family: 'Open Sans',
																				sans-serif;
																			  font-size: 14px;
																			  font-weight: 400;
																			  line-height: 1.5;
																			  margin: 0;
																			  margin-bottom: 10px;
																			  margin-right: 15px;
																			  padding-bottom: 0;
																			  padding-left: 0;
																			  padding-right: 0;
																			  padding-top: 0;
																			  text-align: right;
																			"
																		  >
																			(%v →
																			%v)
																		  </p>
																		</th>
																	  </tr>
																	</tbody>
																  </table>
																</th>
															  </tr>
															</tbody>
														  </table>
														  <table
															class="row"
															style="
															  border-collapse: collapse;
															  border-spacing: 0;
															  display: table;
															  padding: 0;
															  padding-bottom: 0;
															  padding-left: 0;
															  padding-right: 0;
															  padding-top: 0;
															  position: relative;
															  text-align: left;
															  vertical-align: top;
															  width: 100%%;
															"
														  >
															<tbody>
															  <tr
																style="
																  padding-bottom: 0;
																  padding-left: 0;
																  padding-right: 0;
																  padding-top: 0;
																  text-align: left;
																  vertical-align: top;
																"
															  >
																<th
																  class="divider-line small-12 large-12 columns first last"
																  style="
																	-moz-hyphens: auto;
																	-webkit-hyphens: auto;
																	margin: 0 auto;
																	border-collapse: collapse !important;
																	border-top: 1px
																	  solid #e7e9ed;
																	color: #3e516c;
																	display: block;
																	font-family: 'Open Sans',
																	  sans-serif;
																	font-size: 16px;
																	font-weight: 400;
																	height: 1px;
																	hyphens: auto;
																	line-height: 1.5;
																	margin: 0 auto;
																	padding: 5px 0;
																	padding-bottom: 16px;
																	padding-left: 0 !important;
																	padding-right: 0 !important;
																	padding-top: 0;
																	text-align: left;
																	vertical-align: top;
																	width: 100%%;
																	word-wrap: break-word;
																  "
																>
																  <table
																	style="
																	  border-collapse: collapse;
																	  border-spacing: 0;
																	  padding-bottom: 0;
																	  padding-left: 0;
																	  padding-right: 0;
																	  padding-top: 0;
																	  text-align: left;
																	  vertical-align: top;
																	  width: 100%%;
																	"
																  >
																	<tbody>
																	  <tr
																		style="
																		  padding-bottom: 0;
																		  padding-left: 0;
																		  padding-right: 0;
																		  padding-top: 0;
																		  text-align: left;
																		  vertical-align: top;
																		"
																	  >
																		<th
																		  style="
																			-moz-hyphens: auto;
																			-webkit-hyphens: auto;
																			margin: 0;
																			border-collapse: collapse !important;
																			color: #3e516c;
																			font-family: 'Open Sans',
																			  sans-serif;
																			font-size: 16px;
																			font-weight: 400;
																			hyphens: auto;
																			line-height: 1.5;
																			margin: 0;
																			padding-bottom: 0;
																			padding-left: 0;
																			padding-right: 0;
																			padding-top: 0;
																			text-align: left;
																			vertical-align: top;
																			word-wrap: break-word;
																		  "
																		></th>
																		<th
																		  class="expander"
																		  style="
																			-moz-hyphens: auto;
																			-webkit-hyphens: auto;
																			margin: 0;
																			border-collapse: collapse !important;
																			color: #3e516c;
																			font-family: 'Open Sans',
																			  sans-serif;
																			font-size: 16px;
																			font-weight: 400;
																			hyphens: auto;
																			line-height: 1.5;
																			margin: 0;
																			padding: 0 !important;
																			padding-bottom: 0;
																			padding-left: 0;
																			padding-right: 0;
																			padding-top: 0;
																			text-align: left;
																			vertical-align: top;
																			visibility: hidden;
																			width: 0;
																			word-wrap: break-word;
																		  "
																		></th>
																	  </tr>
																	</tbody>
																  </table>
																</th>
															  </tr>
															</tbody>
														  </table>
														  <table
															class="row"
															style="
															  border-collapse: collapse;
															  border-spacing: 0;
															  display: table;
															  padding: 0;
															  padding-bottom: 0;
															  padding-left: 0;
															  padding-right: 0;
															  padding-top: 0;
															  position: relative;
															  text-align: left;
															  vertical-align: top;
															  width: 100%%;
															"
														  >
															<tbody>
															  <tr
																style="
																  padding-bottom: 0;
																  padding-left: 0;
																  padding-right: 0;
																  padding-top: 0;
																  text-align: left;
																  vertical-align: top;
																"
															  >
																<th
																  class="small-12 large-6 columns first"
																  style="
																	-moz-hyphens: auto;
																	-webkit-hyphens: auto;
																	margin: 0 auto;
																	border-collapse: collapse !important;
																	color: #3e516c;
																	font-family: 'Open Sans',
																	  sans-serif;
																	font-size: 16px;
																	font-weight: 400;
																	hyphens: auto;
																	line-height: 1.5;
																	margin: 0 auto;
																	padding-bottom: 16px;
																	padding-left: 0 !important;
																	padding-right: 8px;
																	padding-top: 0;
																	text-align: left;
																	vertical-align: top;
																	width: 50%%;
																	word-wrap: break-word;
																  "
																>
																  <table
																	style="
																	  border-collapse: collapse;
																	  border-spacing: 0;
																	  padding-bottom: 0;
																	  padding-left: 0;
																	  padding-right: 0;
																	  padding-top: 0;
																	  text-align: left;
																	  vertical-align: top;
																	  width: 100%%;
																	"
																  >
																	<tbody>
																	  <tr
																		style="
																		  padding-bottom: 0;
																		  padding-left: 0;
																		  padding-right: 0;
																		  padding-top: 0;
																		  text-align: left;
																		  vertical-align: top;
																		"
																	  >
																		<th
																		  style="
																			-moz-hyphens: auto;
																			-webkit-hyphens: auto;
																			margin: 0;
																			border-collapse: collapse !important;
																			color: #3e516c;
																			font-family: 'Open Sans',
																			  sans-serif;
																			font-size: 16px;
																			font-weight: 400;
																			hyphens: auto;
																			line-height: 1.5;
																			margin: 0;
																			padding-bottom: 0;
																			padding-left: 0;
																			padding-right: 0;
																			padding-top: 0;
																			text-align: left;
																			vertical-align: top;
																			word-wrap: break-word;
																		  "
																		>
																		  <p
																			class="text-left margin-left h7"
																			style="
																			  margin: 0;
																			  margin-bottom: 10px;
																			  color: #3e516c;
																			  font-family: 'Open Sans',
																				sans-serif;
																			  font-size: 14px;
																			  font-weight: 400;
																			  line-height: 1.5;
																			  margin: 0;
																			  margin-bottom: 10px;
																			  margin-left: 15px;
																			  padding-bottom: 0;
																			  padding-left: 0;
																			  padding-right: 0;
																			  padding-top: 0;
																			  text-align: left;
																			"
																		  >
																			%v
																			is
																			<span
																			  class="text-bold"
																			  style="
																				font-weight: 700;
																			  "
																			  >%v</span
																			>
																		  </p>
																		</th>
																	  </tr>
																	</tbody>
																  </table>
																</th>
																<th
																  class="small-12 large-6 columns last"
																  style="
																	-moz-hyphens: auto;
																	-webkit-hyphens: auto;
																	margin: 0 auto;
																	border-collapse: collapse !important;
																	color: #3e516c;
																	font-family: 'Open Sans',
																	  sans-serif;
																	font-size: 16px;
																	font-weight: 400;
																	hyphens: auto;
																	line-height: 1.5;
																	margin: 0 auto;
																	padding-bottom: 16px;
																	padding-left: 8px;
																	padding-right: 0 !important;
																	padding-top: 0;
																	text-align: left;
																	vertical-align: top;
																	width: 50%%;
																	word-wrap: break-word;
																  "
																>
																  <table
																	style="
																	  border-collapse: collapse;
																	  border-spacing: 0;
																	  padding-bottom: 0;
																	  padding-left: 0;
																	  padding-right: 0;
																	  padding-top: 0;
																	  text-align: left;
																	  vertical-align: top;
																	  width: 100%%;
																	"
																  >
																	<tbody>
																	  <tr
																		style="
																		  padding-bottom: 0;
																		  padding-left: 0;
																		  padding-right: 0;
																		  padding-top: 0;
																		  text-align: left;
																		  vertical-align: top;
																		"
																	  >
																		<th
																		  style="
																			-moz-hyphens: auto;
																			-webkit-hyphens: auto;
																			margin: 0;
																			border-collapse: collapse !important;
																			color: #3e516c;
																			font-family: 'Open Sans',
																			  sans-serif;
																			font-size: 16px;
																			font-weight: 400;
																			hyphens: auto;
																			line-height: 1.5;
																			margin: 0;
																			padding-bottom: 0;
																			padding-left: 0;
																			padding-right: 0;
																			padding-top: 0;
																			text-align: left;
																			vertical-align: top;
																			word-wrap: break-word;
																		  "
																		>
																		  <p
																			class="text-right margin-right h7"
																			style="
																			  margin: 0;
																			  margin-bottom: 10px;
																			  color: #3e516c;
																			  font-family: 'Open Sans',
																				sans-serif;
																			  font-size: 14px;
																			  font-weight: 400;
																			  line-height: 1.5;
																			  margin: 0;
																			  margin-bottom: 10px;
																			  margin-right: 15px;
																			  padding-bottom: 0;
																			  padding-left: 0;
																			  padding-right: 0;
																			  padding-top: 0;
																			  text-align: right;
																			"
																		  >
																			%v
																		  </p>
																		</th>
																	  </tr>
																	</tbody>
																  </table>
																</th>
															  </tr>
															</tbody>
														  </table>
														  <table
															class="row"
															style="
															  border-collapse: collapse;
															  border-spacing: 0;
															  display: table;
															  padding: 0;
															  padding-bottom: 0;
															  padding-left: 0;
															  padding-right: 0;
															  padding-top: 0;
															  position: relative;
															  text-align: left;
															  vertical-align: top;
															  width: 100%%;
															"
														  >
															<tbody>
															  <tr
																style="
																  padding-bottom: 0;
																  padding-left: 0;
																  padding-right: 0;
																  padding-top: 0;
																  text-align: left;
																  vertical-align: top;
																"
															  >
																<th
																  class="divider-line-light small-12 large-11 columns first last"
																  style="
																	-moz-hyphens: auto;
																	-webkit-hyphens: auto;
																	margin: 0 auto;
																	border-collapse: collapse !important;
																	border-top: 0.4px
																	  solid #e7e9ed;
																	color: #3e516c;
																	display: block;
																	font-family: 'Open Sans',
																	  sans-serif;
																	font-size: 16px;
																	font-weight: 400;
																	height: 1px;
																	hyphens: auto;
																	line-height: 1.5;
																	margin: 0 auto;
																	padding: 5px 0;
																	padding-bottom: 16px;
																	padding-left: 0 !important;
																	padding-right: 0 !important;
																	padding-top: 0;
																	text-align: left;
																	vertical-align: top;
																	width: 91.66667%%;
																	word-wrap: break-word;
																  "
																>
																  <table
																	style="
																	  border-collapse: collapse;
																	  border-spacing: 0;
																	  padding-bottom: 0;
																	  padding-left: 0;
																	  padding-right: 0;
																	  padding-top: 0;
																	  text-align: left;
																	  vertical-align: top;
																	  width: 100%%;
																	"
																  >
																	<tbody>
																	  <tr
																		style="
																		  padding-bottom: 0;
																		  padding-left: 0;
																		  padding-right: 0;
																		  padding-top: 0;
																		  text-align: left;
																		  vertical-align: top;
																		"
																	  >
																		<th
																		  style="
																			-moz-hyphens: auto;
																			-webkit-hyphens: auto;
																			margin: 0;
																			border-collapse: collapse !important;
																			color: #3e516c;
																			font-family: 'Open Sans',
																			  sans-serif;
																			font-size: 16px;
																			font-weight: 400;
																			hyphens: auto;
																			line-height: 1.5;
																			margin: 0;
																			padding-bottom: 0;
																			padding-left: 0;
																			padding-right: 0;
																			padding-top: 0;
																			text-align: left;
																			vertical-align: top;
																			word-wrap: break-word;
																		  "
																		></th>
																	  </tr>
																	</tbody>
																  </table>
																</th>
															  </tr>
															</tbody>
														  </table>
														  <table
															class="row"
															style="
															  border-collapse: collapse;
															  border-spacing: 0;
															  display: table;
															  padding: 0;
															  padding-bottom: 0;
															  padding-left: 0;
															  padding-right: 0;
															  padding-top: 0;
															  position: relative;
															  text-align: left;
															  vertical-align: top;
															  width: 100%%;
															"
														  >
															<tbody>
															  <tr
																style="
																  padding-bottom: 0;
																  padding-left: 0;
																  padding-right: 0;
																  padding-top: 0;
																  text-align: left;
																  vertical-align: top;
																"
															  >
																<th
																  class="small-12 large-10 columns first"
																  style="
																	-moz-hyphens: auto;
																	-webkit-hyphens: auto;
																	margin: 0 auto;
																	border-collapse: collapse !important;
																	color: #3e516c;
																	font-family: 'Open Sans',
																	  sans-serif;
																	font-size: 16px;
																	font-weight: 400;
																	hyphens: auto;
																	line-height: 1.5;
																	margin: 0 auto;
																	padding-bottom: 16px;
																	padding-left: 0 !important;
																	padding-right: 8px;
																	padding-top: 0;
																	text-align: left;
																	vertical-align: top;
																	width: 83.33333%%;
																	word-wrap: break-word;
																  "
																>
																  <table
																	style="
																	  border-collapse: collapse;
																	  border-spacing: 0;
																	  padding-bottom: 0;
																	  padding-left: 0;
																	  padding-right: 0;
																	  padding-top: 0;
																	  text-align: left;
																	  vertical-align: top;
																	  width: 100%%;
																	"
																  >
																	<tbody>
																	  <tr
																		style="
																		  padding-bottom: 0;
																		  padding-left: 0;
																		  padding-right: 0;
																		  padding-top: 0;
																		  text-align: left;
																		  vertical-align: top;
																		"
																	  >
																		<th
																		  style="
																			-moz-hyphens: auto;
																			-webkit-hyphens: auto;
																			margin: 0;
																			border-collapse: collapse !important;
																			color: #3e516c;
																			font-family: 'Open Sans',
																			  sans-serif;
																			font-size: 16px;
																			font-weight: 400;
																			hyphens: auto;
																			line-height: 1.5;
																			margin: 0;
																			padding-bottom: 0;
																			padding-left: 0;
																			padding-right: 0;
																			padding-top: 0;
																			text-align: left;
																			vertical-align: top;
																			word-wrap: break-word;
																		  "
																		>
																		  <p
																			class="text-left margin-left h7"
																			style="
																			  margin: 0;
																			  margin-bottom: 10px;
																			  color: #3e516c;
																			  font-family: 'Open Sans',
																				sans-serif;
																			  font-size: 14px;
																			  font-weight: 400;
																			  line-height: 1.5;
																			  margin: 0;
																			  margin-bottom: 10px;
																			  margin-left: 15px;
																			  padding-bottom: 0;
																			  padding-left: 0;
																			  padding-right: 0;
																			  padding-top: 0;
																			  text-align: left;
																			"
																		  >
																			%v
																			is
																			<span
																			  class="text-bold"
																			  style="
																				font-weight: 700;
																			  "
																			  >%v</span
																			>
																		  </p>
																		</th>
																	  </tr>
																	</tbody>
																  </table>
																</th>
																<th
																  class="small-12 large-6 columns last"
																  style="
																	-moz-hyphens: auto;
																	-webkit-hyphens: auto;
																	margin: 0 auto;
																	border-collapse: collapse !important;
																	color: #3e516c;
																	font-family: 'Open Sans',
																	  sans-serif;
																	font-size: 16px;
																	font-weight: 400;
																	hyphens: auto;
																	line-height: 1.5;
																	margin: 0 auto;
																	padding-bottom: 16px;
																	padding-left: 8px;
																	padding-right: 0 !important;
																	padding-top: 0;
																	text-align: left;
																	vertical-align: top;
																	width: 50%%;
																	word-wrap: break-word;
																  "
																>
																  <table
																	style="
																	  border-collapse: collapse;
																	  border-spacing: 0;
																	  padding-bottom: 0;
																	  padding-left: 0;
																	  padding-right: 0;
																	  padding-top: 0;
																	  text-align: left;
																	  vertical-align: top;
																	  width: 100%%;
																	"
																  >
																	<tbody>
																	  <tr
																		style="
																		  padding-bottom: 0;
																		  padding-left: 0;
																		  padding-right: 0;
																		  padding-top: 0;
																		  text-align: left;
																		  vertical-align: top;
																		"
																	  >
																		<th
																		  style="
																			-moz-hyphens: auto;
																			-webkit-hyphens: auto;
																			margin: 0;
																			border-collapse: collapse !important;
																			color: #3e516c;
																			font-family: 'Open Sans',
																			  sans-serif;
																			font-size: 16px;
																			font-weight: 400;
																			hyphens: auto;
																			line-height: 1.5;
																			margin: 0;
																			padding-bottom: 0;
																			padding-left: 0;
																			padding-right: 0;
																			padding-top: 0;
																			text-align: left;
																			vertical-align: top;
																			word-wrap: break-word;
																		  "
																		>
																		  <p
																			class="text-right margin-right h7"
																			style="
																			  margin: 0;
																			  margin-bottom: 10px;
																			  color: #3e516c;
																			  font-family: 'Open Sans',
																				sans-serif;
																			  font-size: 14px;
																			  font-weight: 400;
																			  line-height: 1.5;
																			  margin: 0;
																			  margin-bottom: 10px;
																			  margin-right: 15px;
																			  padding-bottom: 0;
																			  padding-left: 0;
																			  padding-right: 0;
																			  padding-top: 0;
																			  text-align: right;
																			"
																		  >
																			%v
																		  </p>
																		</th>
																	  </tr>
																	</tbody>
																  </table>
																</th>
															  </tr>
															</tbody>
														  </table>
														  <table
															class="row"
															style="
															  border-collapse: collapse;
															  border-spacing: 0;
															  display: table;
															  padding: 0;
															  padding-bottom: 0;
															  padding-left: 0;
															  padding-right: 0;
															  padding-top: 0;
															  position: relative;
															  text-align: left;
															  vertical-align: top;
															  width: 100%%;
															"
														  >
															<tbody>
															  <tr
																style="
																  padding-bottom: 0;
																  padding-left: 0;
																  padding-right: 0;
																  padding-top: 0;
																  text-align: left;
																  vertical-align: top;
																"
															  >
																<th
																  class="divider-line-light small-12 large-11 columns first last"
																  style="
																	-moz-hyphens: auto;
																	-webkit-hyphens: auto;
																	margin: 0 auto;
																	border-collapse: collapse !important;
																	border-top: 0.4px
																	  solid #e7e9ed;
																	color: #3e516c;
																	display: block;
																	font-family: 'Open Sans',
																	  sans-serif;
																	font-size: 16px;
																	font-weight: 400;
																	height: 1px;
																	hyphens: auto;
																	line-height: 1.5;
																	margin: 0 auto;
																	padding: 5px 0;
																	padding-bottom: 16px;
																	padding-left: 0 !important;
																	padding-right: 0 !important;
																	padding-top: 0;
																	text-align: left;
																	vertical-align: top;
																	width: 91.66667%%;
																	word-wrap: break-word;
																  "
																>
																  <table
																	style="
																	  border-collapse: collapse;
																	  border-spacing: 0;
																	  padding-bottom: 0;
																	  padding-left: 0;
																	  padding-right: 0;
																	  padding-top: 0;
																	  text-align: left;
																	  vertical-align: top;
																	  width: 100%%;
																	"
																  >
																	<tbody>
																	  <tr
																		style="
																		  padding-bottom: 0;
																		  padding-left: 0;
																		  padding-right: 0;
																		  padding-top: 0;
																		  text-align: left;
																		  vertical-align: top;
																		"
																	  >
																		<th
																		  style="
																			-moz-hyphens: auto;
																			-webkit-hyphens: auto;
																			margin: 0;
																			border-collapse: collapse !important;
																			color: #3e516c;
																			font-family: 'Open Sans',
																			  sans-serif;
																			font-size: 16px;
																			font-weight: 400;
																			hyphens: auto;
																			line-height: 1.5;
																			margin: 0;
																			padding-bottom: 0;
																			padding-left: 0;
																			padding-right: 0;
																			padding-top: 0;
																			text-align: left;
																			vertical-align: top;
																			word-wrap: break-word;
																		  "
																		></th>
																	  </tr>
																	</tbody>
																  </table>
																</th>
															  </tr>
															</tbody>
														  </table>
														  <table
															class="row"
															style="
															  border-collapse: collapse;
															  border-spacing: 0;
															  display: table;
															  padding: 0;
															  padding-bottom: 0;
															  padding-left: 0;
															  padding-right: 0;
															  padding-top: 0;
															  position: relative;
															  text-align: left;
															  vertical-align: top;
															  width: 100%%;
															"
														  >
															<tbody>
															  <tr
																style="
																  padding-bottom: 0;
																  padding-left: 0;
																  padding-right: 0;
																  padding-top: 0;
																  text-align: left;
																  vertical-align: top;
																"
															  >
																<th
																  class="small-12 large-10 columns first"
																  style="
																	-moz-hyphens: auto;
																	-webkit-hyphens: auto;
																	margin: 0 auto;
																	border-collapse: collapse !important;
																	color: #3e516c;
																	font-family: 'Open Sans',
																	  sans-serif;
																	font-size: 16px;
																	font-weight: 400;
																	hyphens: auto;
																	line-height: 1.5;
																	margin: 0 auto;
																	padding-bottom: 16px;
																	padding-left: 0 !important;
																	padding-right: 8px;
																	padding-top: 0;
																	text-align: left;
																	vertical-align: top;
																	width: 83.33333%%;
																	word-wrap: break-word;
																  "
																>
																  <table
																	style="
																	  border-collapse: collapse;
																	  border-spacing: 0;
																	  padding-bottom: 0;
																	  padding-left: 0;
																	  padding-right: 0;
																	  padding-top: 0;
																	  text-align: left;
																	  vertical-align: top;
																	  width: 100%%;
																	"
																  >
																	<tbody>
																	  <tr
																		style="
																		  padding-bottom: 0;
																		  padding-left: 0;
																		  padding-right: 0;
																		  padding-top: 0;
																		  text-align: left;
																		  vertical-align: top;
																		"
																	  >
																		<th
																		  style="
																			-moz-hyphens: auto;
																			-webkit-hyphens: auto;
																			margin: 0;
																			border-collapse: collapse !important;
																			color: #3e516c;
																			font-family: 'Open Sans',
																			  sans-serif;
																			font-size: 16px;
																			font-weight: 400;
																			hyphens: auto;
																			line-height: 1.5;
																			margin: 0;
																			padding-bottom: 0;
																			padding-left: 0;
																			padding-right: 0;
																			padding-top: 0;
																			text-align: left;
																			vertical-align: top;
																			word-wrap: break-word;
																		  "
																		>
																		  <p
																			class="text-left margin-left h7"
																			style="
																			  margin: 0;
																			  margin-bottom: 10px;
																			  color: #3e516c;
																			  font-family: 'Open Sans',
																				sans-serif;
																			  font-size: 14px;
																			  font-weight: 400;
																			  line-height: 1.5;
																			  margin: 0;
																			  margin-bottom: 10px;
																			  margin-left: 15px;
																			  padding-bottom: 0;
																			  padding-left: 0;
																			  padding-right: 0;
																			  padding-top: 0;
																			  text-align: left;
																			"
																		  >
																			%v
																			is
																			<span
																			  class="text-bold"
																			  style="
																				font-weight: 700;
																			  "
																			  >%v</span
																			>
																		  </p>
																		</th>
																	  </tr>
																	</tbody>
																  </table>
																</th>
																<th
																  class="small-12 large-6 columns last"
																  style="
																	-moz-hyphens: auto;
																	-webkit-hyphens: auto;
																	margin: 0 auto;
																	border-collapse: collapse !important;
																	color: #3e516c;
																	font-family: 'Open Sans',
																	  sans-serif;
																	font-size: 16px;
																	font-weight: 400;
																	hyphens: auto;
																	line-height: 1.5;
																	margin: 0 auto;
																	padding-bottom: 16px;
																	padding-left: 8px;
																	padding-right: 0 !important;
																	padding-top: 0;
																	text-align: left;
																	vertical-align: top;
																	width: 50%%;
																	word-wrap: break-word;
																  "
																>
																  <table
																	style="
																	  border-collapse: collapse;
																	  border-spacing: 0;
																	  padding-bottom: 0;
																	  padding-left: 0;
																	  padding-right: 0;
																	  padding-top: 0;
																	  text-align: left;
																	  vertical-align: top;
																	  width: 100%%;
																	"
																  >
																	<tbody>
																	  <tr
																		style="
																		  padding-bottom: 0;
																		  padding-left: 0;
																		  padding-right: 0;
																		  padding-top: 0;
																		  text-align: left;
																		  vertical-align: top;
																		"
																	  >
																		<th
																		  style="
																			-moz-hyphens: auto;
																			-webkit-hyphens: auto;
																			margin: 0;
																			border-collapse: collapse !important;
																			color: #3e516c;
																			font-family: 'Open Sans',
																			  sans-serif;
																			font-size: 16px;
																			font-weight: 400;
																			hyphens: auto;
																			line-height: 1.5;
																			margin: 0;
																			padding-bottom: 0;
																			padding-left: 0;
																			padding-right: 0;
																			padding-top: 0;
																			text-align: left;
																			vertical-align: top;
																			word-wrap: break-word;
																		  "
																		>
																		  <p
																			class="text-right margin-right h7"
																			style="
																			  margin: 0;
																			  margin-bottom: 10px;
																			  color: #3e516c;
																			  font-family: 'Open Sans',
																				sans-serif;
																			  font-size: 14px;
																			  font-weight: 400;
																			  line-height: 1.5;
																			  margin: 0;
																			  margin-bottom: 10px;
																			  margin-right: 15px;
																			  padding-bottom: 0;
																			  padding-left: 0;
																			  padding-right: 0;
																			  padding-top: 0;
																			  text-align: right;
																			"
																		  >
																			%v
																		  </p>
																		</th>
																	  </tr>
																	</tbody>
																  </table>
																</th>
															  </tr>
															</tbody>
														  </table>
														</th>
													  </tr>
													</tbody>
												  </table>
												</th>
											  </tr>
											</tbody>
										  </table>
										  <table
											align="center"
											class="spacer float-center"
											style="
											  margin: 0 auto;
											  border-collapse: collapse;
											  border-spacing: 0;
											  float: none;
											  margin: 0 auto;
											  padding-bottom: 0;
											  padding-left: 0;
											  padding-right: 0;
											  padding-top: 0;
											  text-align: center;
											  vertical-align: top;
											  width: 100%%;
											"
										  >
											<tbody>
											  <tr
												style="
												  padding-bottom: 0;
												  padding-left: 0;
												  padding-right: 0;
												  padding-top: 0;
												  text-align: left;
												  vertical-align: top;
												"
											  >
												<td
												  height="16"
												  style="
													-moz-hyphens: auto;
													-webkit-hyphens: auto;
													margin: 0;
													border-collapse: collapse !important;
													color: #3e516c;
													font-family: 'Open Sans',
													  sans-serif;
													font-size: 16px;
													font-weight: 400;
													hyphens: auto;
													line-height: 16px;
													margin: 0;
													mso-line-height-rule: exactly;
													padding-bottom: 0;
													padding-left: 0;
													padding-right: 0;
													padding-top: 0;
													text-align: left;
													vertical-align: top;
													word-wrap: break-word;
												  "
												>
												  &nbsp;
												</td>
											  </tr>
											</tbody>
										  </table>
										</center>
										<table
										  class="spacer"
										  style="
											border-collapse: collapse;
											border-spacing: 0;
											padding-bottom: 0;
											padding-left: 0;
											padding-right: 0;
											padding-top: 0;
											text-align: left;
											vertical-align: top;
											width: 100%%;
										  "
										>
										  <tbody>
											<tr
											  style="
												padding-bottom: 0;
												padding-left: 0;
												padding-right: 0;
												padding-top: 0;
												text-align: left;
												vertical-align: top;
											  "
											>
											  <td
												height="16"
												style="
												  -moz-hyphens: auto;
												  -webkit-hyphens: auto;
												  margin: 0;
												  border-collapse: collapse !important;
												  color: #3e516c;
												  font-family: 'Open Sans',
													sans-serif;
												  font-size: 16px;
												  font-weight: 400;
												  hyphens: auto;
												  line-height: 16px;
												  margin: 0;
												  mso-line-height-rule: exactly;
												  padding-bottom: 0;
												  padding-left: 0;
												  padding-right: 0;
												  padding-top: 0;
												  text-align: left;
												  vertical-align: top;
												  word-wrap: break-word;
												"
											  >
												&nbsp;
											  </td>
											</tr>
										  </tbody>
										</table>
										<center
										  style="min-width: 338.67px; width: 100%%"
										>
										  <a
											class="primary-btn small float-center"
											target="_blank"
											href="https://app.factors.ai/login"
											align="center"
											style="
											  background-color: #1890ff;
											  border-radius: 4px;
											  color: #fff !important;
											  font-family: 'Open Sans', sans-serif;
											  font-size: 14px;
											  font-weight: 400;
											  line-height: 1.5;
											  min-width: 150px;
											  padding: 8px 20px;
											  text-align: center;
											  text-decoration: none;
											"
											>Show full report</a
										  >
										</center>
									  </th>
									</tr>
								  </tbody>
								</table>
							  </th>
							  <th
								class="small-12 large-2 columns last"
								style="
								  -moz-hyphens: auto;
								  -webkit-hyphens: auto;
								  margin: 0 auto;
								  border-collapse: collapse !important;
								  color: #3e516c;
								  font-family: 'Open Sans', sans-serif;
								  font-size: 16px;
								  font-weight: 400;
								  hyphens: auto;
								  line-height: 1.5;
								  margin: 0 auto;
								  padding-bottom: 16px;
								  padding-left: 8px;
								  padding-right: 16px;
								  padding-top: 0;
								  text-align: left;
								  vertical-align: top;
								  width: 80.67px;
								  word-wrap: break-word;
								"
							  >
								<table
								  style="
									border-collapse: collapse;
									border-spacing: 0;
									padding-bottom: 0;
									padding-left: 0;
									padding-right: 0;
									padding-top: 0;
									text-align: left;
									vertical-align: top;
									width: 100%%;
								  "
								>
								  <tbody>
									<tr
									  style="
										padding-bottom: 0;
										padding-left: 0;
										padding-right: 0;
										padding-top: 0;
										text-align: left;
										vertical-align: top;
									  "
									>
									  <th
										style="
										  -moz-hyphens: auto;
										  -webkit-hyphens: auto;
										  margin: 0;
										  border-collapse: collapse !important;
										  color: #3e516c;
										  font-family: 'Open Sans', sans-serif;
										  font-size: 16px;
										  font-weight: 400;
										  hyphens: auto;
										  line-height: 1.5;
										  margin: 0;
										  padding-bottom: 0;
										  padding-left: 0;
										  padding-right: 0;
										  padding-top: 0;
										  text-align: left;
										  vertical-align: top;
										  word-wrap: break-word;
										"
									  ></th>
									</tr>
								  </tbody>
								</table>
							  </th>
							</tr>
						  </tbody>
						</table>
						<table
						  class="row"
						  style="
							border-collapse: collapse;
							border-spacing: 0;
							display: table;
							padding: 0;
							padding-bottom: 0;
							padding-left: 0;
							padding-right: 0;
							padding-top: 0;
							position: relative;
							text-align: left;
							vertical-align: top;
							width: 100%%;
						  "
						>
						  <tbody>
							<tr
							  style="
								padding-bottom: 0;
								padding-left: 0;
								padding-right: 0;
								padding-top: 0;
								text-align: left;
								vertical-align: top;
							  "
							>
							  <th
								class="small-12 large-2 columns first"
								style="
								  -moz-hyphens: auto;
								  -webkit-hyphens: auto;
								  margin: 0 auto;
								  border-collapse: collapse !important;
								  color: #3e516c;
								  font-family: 'Open Sans', sans-serif;
								  font-size: 16px;
								  font-weight: 400;
								  hyphens: auto;
								  line-height: 1.5;
								  margin: 0 auto;
								  padding-bottom: 16px;
								  padding-left: 16px;
								  padding-right: 8px;
								  padding-top: 0;
								  text-align: left;
								  vertical-align: top;
								  width: 80.67px;
								  word-wrap: break-word;
								"
							  >
								<table
								  style="
									border-collapse: collapse;
									border-spacing: 0;
									padding-bottom: 0;
									padding-left: 0;
									padding-right: 0;
									padding-top: 0;
									text-align: left;
									vertical-align: top;
									width: 100%%;
								  "
								>
								  <tbody>
									<tr
									  style="
										padding-bottom: 0;
										padding-left: 0;
										padding-right: 0;
										padding-top: 0;
										text-align: left;
										vertical-align: top;
									  "
									>
									  <th
										style="
										  -moz-hyphens: auto;
										  -webkit-hyphens: auto;
										  margin: 0;
										  border-collapse: collapse !important;
										  color: #3e516c;
										  font-family: 'Open Sans', sans-serif;
										  font-size: 16px;
										  font-weight: 400;
										  hyphens: auto;
										  line-height: 1.5;
										  margin: 0;
										  padding-bottom: 0;
										  padding-left: 0;
										  padding-right: 0;
										  padding-top: 0;
										  text-align: left;
										  vertical-align: top;
										  word-wrap: break-word;
										"
									  ></th>
									</tr>
								  </tbody>
								</table>
							  </th>
							  <th
								class="small-12 large-8 columns"
								style="
								  -moz-hyphens: auto;
								  -webkit-hyphens: auto;
								  margin: 0 auto;
								  border-collapse: collapse !important;
								  color: #3e516c;
								  font-family: 'Open Sans', sans-serif;
								  font-size: 16px;
								  font-weight: 400;
								  hyphens: auto;
								  line-height: 1.5;
								  margin: 0 auto;
								  padding-bottom: 16px;
								  padding-left: 8px;
								  padding-right: 8px;
								  padding-top: 0;
								  text-align: left;
								  vertical-align: top;
								  width: 370.67px;
								  word-wrap: break-word;
								"
							  >
								<table
								  style="
									border-collapse: collapse;
									border-spacing: 0;
									padding-bottom: 0;
									padding-left: 0;
									padding-right: 0;
									padding-top: 0;
									text-align: left;
									vertical-align: top;
									width: 100%%;
								  "
								>
								  <tbody>
									<tr
									  style="
										padding-bottom: 0;
										padding-left: 0;
										padding-right: 0;
										padding-top: 0;
										text-align: left;
										vertical-align: top;
									  "
									>
									  <th
										style="
										  -moz-hyphens: auto;
										  -webkit-hyphens: auto;
										  margin: 0;
										  border-collapse: collapse !important;
										  color: #3e516c;
										  font-family: 'Open Sans', sans-serif;
										  font-size: 16px;
										  font-weight: 400;
										  hyphens: auto;
										  line-height: 1.5;
										  margin: 0;
										  padding-bottom: 0;
										  padding-left: 0;
										  padding-right: 0;
										  padding-top: 0;
										  text-align: left;
										  vertical-align: top;
										  word-wrap: break-word;
										"
									  >
										<table
										  class="spacer"
										  style="
											border-collapse: collapse;
											border-spacing: 0;
											padding-bottom: 0;
											padding-left: 0;
											padding-right: 0;
											padding-top: 0;
											text-align: left;
											vertical-align: top;
											width: 100%%;
										  "
										>
										  <tbody>
											<tr
											  style="
												padding-bottom: 0;
												padding-left: 0;
												padding-right: 0;
												padding-top: 0;
												text-align: left;
												vertical-align: top;
											  "
											>
											  <td
												height="64"
												style="
												  -moz-hyphens: auto;
												  -webkit-hyphens: auto;
												  margin: 0;
												  border-collapse: collapse !important;
												  color: #3e516c;
												  font-family: 'Open Sans',
													sans-serif;
												  font-size: 64px;
												  font-weight: 400;
												  hyphens: auto;
												  line-height: 64px;
												  margin: 0;
												  mso-line-height-rule: exactly;
												  padding-bottom: 0;
												  padding-left: 0;
												  padding-right: 0;
												  padding-top: 0;
												  text-align: left;
												  vertical-align: top;
												  word-wrap: break-word;
												"
											  >
												&nbsp;
											  </td>
											</tr>
										  </tbody>
										</table>
										<p
										  class="text-center h7 medium-grey"
										  style="
											margin: 0;
											margin-bottom: 10px;
											color: #8692a3;
											font-family: 'Open Sans', sans-serif;
											font-size: 14px;
											font-weight: 400;
											line-height: 1.5;
											margin: 0;
											margin-bottom: 10px;
											padding-bottom: 0;
											padding-left: 0;
											padding-right: 0;
											padding-top: 0;
											text-align: center;
										  "
										>
										  If you have any questions, drop us a
										  message using Intercom Chat inside Factors
										  app or email us at<br /><a
											target="_blank"
											href="mailto:support@factors.ai"
											style="
											  color: #1890ff;
											  font-family: 'Open Sans', sans-serif;
											  font-weight: 400;
											  line-height: 1.5;
											  padding: 0;
											  text-align: left;
											  text-decoration: none;
											"
											>support@factors.ai</a
										  >
										</p>
									  </th>
									</tr>
								  </tbody>
								</table>
							  </th>
							  <th
								class="small-12 large-2 columns last"
								style="
								  -moz-hyphens: auto;
								  -webkit-hyphens: auto;
								  margin: 0 auto;
								  border-collapse: collapse !important;
								  color: #3e516c;
								  font-family: 'Open Sans', sans-serif;
								  font-size: 16px;
								  font-weight: 400;
								  hyphens: auto;
								  line-height: 1.5;
								  margin: 0 auto;
								  padding-bottom: 16px;
								  padding-left: 8px;
								  padding-right: 16px;
								  padding-top: 0;
								  text-align: left;
								  vertical-align: top;
								  width: 80.67px;
								  word-wrap: break-word;
								"
							  >
								<table
								  style="
									border-collapse: collapse;
									border-spacing: 0;
									padding-bottom: 0;
									padding-left: 0;
									padding-right: 0;
									padding-top: 0;
									text-align: left;
									vertical-align: top;
									width: 100%%;
								  "
								>
								  <tbody>
									<tr
									  style="
										padding-bottom: 0;
										padding-left: 0;
										padding-right: 0;
										padding-top: 0;
										text-align: left;
										vertical-align: top;
									  "
									>
									  <th
										style="
										  -moz-hyphens: auto;
										  -webkit-hyphens: auto;
										  margin: 0;
										  border-collapse: collapse !important;
										  color: #3e516c;
										  font-family: 'Open Sans', sans-serif;
										  font-size: 16px;
										  font-weight: 400;
										  hyphens: auto;
										  line-height: 1.5;
										  margin: 0;
										  padding-bottom: 0;
										  padding-left: 0;
										  padding-right: 0;
										  padding-top: 0;
										  text-align: left;
										  vertical-align: top;
										  word-wrap: break-word;
										"
									  ></th>
									</tr>
								  </tbody>
								</table>
							  </th>
							</tr>
						  </tbody>
						</table>
						<table
						  class="row"
						  style="
							border-collapse: collapse;
							border-spacing: 0;
							display: table;
							padding: 0;
							padding-bottom: 0;
							padding-left: 0;
							padding-right: 0;
							padding-top: 0;
							position: relative;
							text-align: left;
							vertical-align: top;
							width: 100%%;
						  "
						>
						  <tbody>
							<tr
							  style="
								padding-bottom: 0;
								padding-left: 0;
								padding-right: 0;
								padding-top: 0;
								text-align: left;
								vertical-align: top;
							  "
							>
							  <th
								class="small-12 large-2 columns first"
								style="
								  -moz-hyphens: auto;
								  -webkit-hyphens: auto;
								  margin: 0 auto;
								  border-collapse: collapse !important;
								  color: #3e516c;
								  font-family: 'Open Sans', sans-serif;
								  font-size: 16px;
								  font-weight: 400;
								  hyphens: auto;
								  line-height: 1.5;
								  margin: 0 auto;
								  padding-bottom: 16px;
								  padding-left: 16px;
								  padding-right: 8px;
								  padding-top: 0;
								  text-align: left;
								  vertical-align: top;
								  width: 80.67px;
								  word-wrap: break-word;
								"
							  >
								<table
								  style="
									border-collapse: collapse;
									border-spacing: 0;
									padding-bottom: 0;
									padding-left: 0;
									padding-right: 0;
									padding-top: 0;
									text-align: left;
									vertical-align: top;
									width: 100%%;
								  "
								>
								  <tbody>
									<tr
									  style="
										padding-bottom: 0;
										padding-left: 0;
										padding-right: 0;
										padding-top: 0;
										text-align: left;
										vertical-align: top;
									  "
									>
									  <th
										style="
										  -moz-hyphens: auto;
										  -webkit-hyphens: auto;
										  margin: 0;
										  border-collapse: collapse !important;
										  color: #3e516c;
										  font-family: 'Open Sans', sans-serif;
										  font-size: 16px;
										  font-weight: 400;
										  hyphens: auto;
										  line-height: 1.5;
										  margin: 0;
										  padding-bottom: 0;
										  padding-left: 0;
										  padding-right: 0;
										  padding-top: 0;
										  text-align: left;
										  vertical-align: top;
										  word-wrap: break-word;
										"
									  ></th>
									</tr>
								  </tbody>
								</table>
							  </th>
							  <th
								class="small-12 large-8 columns"
								style="
								  -moz-hyphens: auto;
								  -webkit-hyphens: auto;
								  margin: 0 auto;
								  border-collapse: collapse !important;
								  color: #3e516c;
								  font-family: 'Open Sans', sans-serif;
								  font-size: 16px;
								  font-weight: 400;
								  hyphens: auto;
								  line-height: 1.5;
								  margin: 0 auto;
								  padding-bottom: 16px;
								  padding-left: 8px;
								  padding-right: 8px;
								  padding-top: 0;
								  text-align: left;
								  vertical-align: top;
								  width: 370.67px;
								  word-wrap: break-word;
								"
							  >
								<table
								  style="
									border-collapse: collapse;
									border-spacing: 0;
									padding-bottom: 0;
									padding-left: 0;
									padding-right: 0;
									padding-top: 0;
									text-align: left;
									vertical-align: top;
									width: 100%%;
								  "
								>
								  <tbody>
									<tr
									  style="
										padding-bottom: 0;
										padding-left: 0;
										padding-right: 0;
										padding-top: 0;
										text-align: left;
										vertical-align: top;
									  "
									>
									  <th
										style="
										  -moz-hyphens: auto;
										  -webkit-hyphens: auto;
										  margin: 0;
										  border-collapse: collapse !important;
										  color: #3e516c;
										  font-family: 'Open Sans', sans-serif;
										  font-size: 16px;
										  font-weight: 400;
										  hyphens: auto;
										  line-height: 1.5;
										  margin: 0;
										  padding-bottom: 0;
										  padding-left: 0;
										  padding-right: 0;
										  padding-top: 0;
										  text-align: left;
										  vertical-align: top;
										  word-wrap: break-word;
										"
									  >
										<table
										  class="spacer"
										  style="
											border-collapse: collapse;
											border-spacing: 0;
											padding-bottom: 0;
											padding-left: 0;
											padding-right: 0;
											padding-top: 0;
											text-align: left;
											vertical-align: top;
											width: 100%%;
										  "
										>
										  <tbody>
											<tr
											  style="
												padding-bottom: 0;
												padding-left: 0;
												padding-right: 0;
												padding-top: 0;
												text-align: left;
												vertical-align: top;
											  "
											>
											  <td
												height="32"
												style="
												  -moz-hyphens: auto;
												  -webkit-hyphens: auto;
												  margin: 0;
												  border-collapse: collapse !important;
												  color: #3e516c;
												  font-family: 'Open Sans',
													sans-serif;
												  font-size: 32px;
												  font-weight: 400;
												  hyphens: auto;
												  line-height: 32px;
												  margin: 0;
												  mso-line-height-rule: exactly;
												  padding-bottom: 0;
												  padding-left: 0;
												  padding-right: 0;
												  padding-top: 0;
												  text-align: left;
												  vertical-align: top;
												  word-wrap: break-word;
												"
											  >
												&nbsp;
											  </td>
											</tr>
										  </tbody>
										</table>
										<p
										  class="text-center h7"
										  style="
											margin: 0;
											margin-bottom: 10px;
											color: #3e516c;
											font-family: 'Open Sans', sans-serif;
											font-size: 14px;
											font-weight: 400;
											line-height: 1.5;
											margin: 0;
											margin-bottom: 10px;
											padding-bottom: 0;
											padding-left: 0;
											padding-right: 0;
											padding-top: 0;
											text-align: center;
										  "
										>
										  Happy analysing,
										</p>
										<p
										  class="text-center h7"
										  style="
											margin: 0;
											margin-bottom: 10px;
											color: #3e516c;
											font-family: 'Open Sans', sans-serif;
											font-size: 14px;
											font-weight: 400;
											line-height: 1.5;
											margin: 0;
											margin-bottom: 10px;
											padding-bottom: 0;
											padding-left: 0;
											padding-right: 0;
											padding-top: 0;
											text-align: center;
										  "
										>
										  Team
										  <a
											target="_blank"
											href="http://factors.ai/"
											style="
											  color: #1890ff;
											  font-family: 'Open Sans', sans-serif;
											  font-weight: 400;
											  line-height: 1.5;
											  padding: 0;
											  text-align: left;
											  text-decoration: none;
											"
											>Factors.AI</a
										  >
										</p>
									  </th>
									</tr>
								  </tbody>
								</table>
							  </th>
							  <th
								class="small-12 large-2 columns last"
								style="
								  -moz-hyphens: auto;
								  -webkit-hyphens: auto;
								  margin: 0 auto;
								  border-collapse: collapse !important;
								  color: #3e516c;
								  font-family: 'Open Sans', sans-serif;
								  font-size: 16px;
								  font-weight: 400;
								  hyphens: auto;
								  line-height: 1.5;
								  margin: 0 auto;
								  padding-bottom: 16px;
								  padding-left: 8px;
								  padding-right: 16px;
								  padding-top: 0;
								  text-align: left;
								  vertical-align: top;
								  width: 80.67px;
								  word-wrap: break-word;
								"
							  >
								<table
								  style="
									border-collapse: collapse;
									border-spacing: 0;
									padding-bottom: 0;
									padding-left: 0;
									padding-right: 0;
									padding-top: 0;
									text-align: left;
									vertical-align: top;
									width: 100%%;
								  "
								>
								  <tbody>
									<tr
									  style="
										padding-bottom: 0;
										padding-left: 0;
										padding-right: 0;
										padding-top: 0;
										text-align: left;
										vertical-align: top;
									  "
									>
									  <th
										style="
										  -moz-hyphens: auto;
										  -webkit-hyphens: auto;
										  margin: 0;
										  border-collapse: collapse !important;
										  color: #3e516c;
										  font-family: 'Open Sans', sans-serif;
										  font-size: 16px;
										  font-weight: 400;
										  hyphens: auto;
										  line-height: 1.5;
										  margin: 0;
										  padding-bottom: 0;
										  padding-left: 0;
										  padding-right: 0;
										  padding-top: 0;
										  text-align: left;
										  vertical-align: top;
										  word-wrap: break-word;
										"
									  ></th>
									</tr>
								  </tbody>
								</table>
							  </th>
							</tr>
						  </tbody>
						</table>
					  </td>
					</tr>
				  </tbody>
				</table>
				<table
				  align="center"
				  class="spacer float-center"
				  style="
					margin: 0 auto;
					border-collapse: collapse;
					border-spacing: 0;
					float: none;
					margin: 0 auto;
					padding-bottom: 0;
					padding-left: 0;
					padding-right: 0;
					padding-top: 0;
					text-align: center;
					vertical-align: top;
					width: 100%%;
				  "
				>
				  <tbody>
					<tr
					  style="
						padding-bottom: 0;
						padding-left: 0;
						padding-right: 0;
						padding-top: 0;
						text-align: left;
						vertical-align: top;
					  "
					>
					  <td
						height="16"
						style="
						  -moz-hyphens: auto;
						  -webkit-hyphens: auto;
						  margin: 0;
						  border-collapse: collapse !important;
						  color: #3e516c;
						  font-family: 'Open Sans', sans-serif;
						  font-size: 16px;
						  font-weight: 400;
						  hyphens: auto;
						  line-height: 16px;
						  margin: 0;
						  mso-line-height-rule: exactly;
						  padding-bottom: 0;
						  padding-left: 0;
						  padding-right: 0;
						  padding-top: 0;
						  text-align: left;
						  vertical-align: top;
						  word-wrap: break-word;
						"
					  >
						&nbsp;
					  </td>
					</tr>
				  </tbody>
				</table>
			  </center>
			</td>
		  </tr>
		</table>
		<!-- prevent Gmail on iOS font size manipulation -->
		<div
		  style="
			display: none;
			white-space: nowrap;
			font: 15px courier;
			line-height: 0;
		  "
		>
		  &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp;
		  &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp;
		  &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp; &nbsp;
		</div>
	  </body>
	</html>
	`, agentName, digestHeader, headerPerc, headerValue1, headerValue2, line1Value1, line1Value2, line1Perc, line2Value1, line2Value2, line2Perc, line3Value1, line3Value2, line3Perc)
	return template
}
