package main

import (
	C "factors/config"
	"factors/model/model"
	"net/http"

	"factors/sdk"
	"factors/util"
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
)

/*
Manual identification of few users.
go run run_manual_identify.go --file_path=<path-to-denormalized-file> --project_id=<projectId>
*/
func main() {

	env := flag.String("env", C.DEVELOPMENT, "")

	memSQLHost := flag.String("memsql_host", C.MemSQLDefaultDBParams.Host, "")
	isPSCHost := flag.Int("memsql_is_psc_host", C.MemSQLDefaultDBParams.IsPSCHost, "")
	memSQLPort := flag.Int("memsql_port", C.MemSQLDefaultDBParams.Port, "")
	memSQLUser := flag.String("memsql_user", C.MemSQLDefaultDBParams.User, "")
	memSQLName := flag.String("memsql_name", C.MemSQLDefaultDBParams.Name, "")
	memSQLPass := flag.String("memsql_pass", C.MemSQLDefaultDBParams.Password, "")
	memSQLCertificate := flag.String("memsql_cert", "", "")
	primaryDatastore := flag.String("primary_datastore", C.DatastoreTypeMemSQL, "Primary datastore type as memsql or postgres")

	redisHost := flag.String("redis_host", "localhost", "")
	redisPort := flag.Int("redis_port", 6379, "")
	redisHostPersistent := flag.String("redis_host_ps", "localhost", "")
	redisPortPersistent := flag.Int("redis_port_ps", 6379, "")
	sentryDSN := flag.String("sentry_dsn", "", "Sentry DSN")

	projectID := flag.Int64("project_id", 2251799817000005, "Project Id.")
	liveRun := flag.Bool("live_run", false, "Run live")
	flag.Parse()
	defer util.NotifyOnPanic("Task#run_manual_identify.", *env)

	taskID := "run_manual_identify"

	if *projectID == 0 {
		log.Error("projectId not provided")
		os.Exit(1)
	}
	config := &C.Configuration{
		AppName: taskID,
		Env:     *env,
		MemSQLInfo: C.DBConf{
			Host:        *memSQLHost,
			IsPSCHost:   *isPSCHost,
			Port:        *memSQLPort,
			User:        *memSQLUser,
			Name:        *memSQLName,
			Password:    *memSQLPass,
			Certificate: *memSQLCertificate,
			AppName:     taskID,
		},
		SentryDSN:           *sentryDSN,
		PrimaryDatastore:    *primaryDatastore,
		RedisHost:           *redisHost,
		RedisPort:           *redisPort,
		RedisHostPersistent: *redisHostPersistent,
		RedisPortPersistent: *redisPortPersistent,
	}

	C.InitConf(config)
	C.InitSentryLogging(config.SentryDSN, config.AppName)

	err := C.InitDB(*config)
	if err != nil {
		log.Error("Failed to initialize DB.")
		os.Exit(1)
	}

	C.InitRedis(config.RedisHost, config.RedisPort)
	C.InitRedisPersistent(config.RedisHostPersistent, config.RedisPortPersistent)

	if *liveRun {
		log.Info("Running live.")
	} else {
		log.Info("Running dry.")
	}

	logCtx := log.WithFields(log.Fields{"project_id": *projectID})

	userIDEmailMap := getUserIdEmailMap()

	for userID, email := range userIDEmailMap {
		if email != "" && userID != "" {
			logCtx.WithFields(log.Fields{"customer_user_id": email, "userID": userID}).Info(
				"Starting to identify user.")
			if *liveRun {
				status, _ := sdk.Identify(*projectID, &sdk.IdentifyPayload{
					UserId: userID, CustomerUserId: email, RequestSource: model.UserSourceWeb}, true)
				if status != http.StatusOK {
					logCtx.WithFields(log.Fields{"customer_user_id": email, "userID": userID}).Error(
						"Failed to identify user.")
				} else {
					logCtx.WithFields(log.Fields{"customer_user_id": email, "userID": userID}).Info(
						"Successfully Identified user from source web.")
				}
			}
		}
	}
}

func getUserIdEmailMap() map[string]string {
	userIDEmailMap := make(map[string]string)
	userIDEmailMap["1773e3c9-98e6-4a0c-8a3a-0c00a6112cef"] = "abm18027@iiml.ac.in"
	userIDEmailMap["77db8c11-2ea7-44ca-aa37-0fbf95d14c42"] = "accounts@nobleqatar.com"
	userIDEmailMap["3e6d8509-543c-45cd-8cca-d07390e419ff"] = "aceonelimit@ace.com"
	userIDEmailMap["0ed13d13-f9dc-426e-a5f8-3f196f168b41"] = "adebanjo.adeleke@konga.com"
	userIDEmailMap["e364bb99-f0f2-44c9-9ac8-18049e8f9ffa"] = "admin@daystolpower.com"
	userIDEmailMap["9cd641f2-c60d-4bc2-bd9e-769c51a1c0de"] = "admin@kenyawriters.com"
	userIDEmailMap["503919c4-1292-4921-9068-ff0bba171f9b"] = "admin@mybeezapp.com"
	userIDEmailMap["90dda389-cd36-4985-860b-84b7aa577f8c"] = "admin@tamashauae.com"
	userIDEmailMap["72afa16e-0502-4cd4-aff5-c2f92f81bc2c"] = "admin@xule.co.ke"
	userIDEmailMap["cc76c230-cb1e-49d7-973d-828a0605e45a"] = "advertise@donkuy.co.za"
	userIDEmailMap["f11fa4ba-82fa-4a3f-aeb9-8e21324a0bda"] = "agustin@visinemapictures.com"
	userIDEmailMap["a72cb40b-f73d-4dbe-969c-6c658ef20cb8"] = "ahmed-allam@elsabaautoservice.com"
	userIDEmailMap["ef1619ff-c8e6-403d-bc46-2ecfa518f5ea"] = "ahmed@babakhabbaz.com"
	userIDEmailMap["a38ab507-add3-4824-a0b5-c76cc28c617f"] = "ahmed@white-tracks.com"
	userIDEmailMap["756f9dd8-253c-471f-954c-6bfdf5a49ca9"] = "aiden@childrenoflife.co.za"
	userIDEmailMap["e57c6a71-5b48-4a50-accd-9645c73d66bf"] = "akinleye.akinsola@thelacaseracompany.com"
	userIDEmailMap["36ecb793-4e6a-4021-8916-4a9b75059cfd"] = "akshay@cssdubai.com"
	userIDEmailMap["a2f2a9cb-7455-4166-ae51-c48bb27c5d83"] = "alex@levetot.co.ke"
	userIDEmailMap["41ade211-b6ea-4e2a-8d9f-c60ca2f15718"] = "aliu.tunde@thelagoscontinental.com"
	userIDEmailMap["30ca799a-e875-4429-8bff-1be6aed38255"] = "amaka@ade.com"
	userIDEmailMap["03999fb9-2649-4548-99a3-bd81f081ae12"] = "amitgupta@financeway.in"
	userIDEmailMap["6b0967a3-59af-4aee-8ab2-d602070a1be9"] = "anandk@goandroytechnologies.com"
	userIDEmailMap["3bfe8337-4296-4a6a-a680-f9e5f0c47389"] = "andrew.cxy@mobvoyage.com"
	userIDEmailMap["c0550902-8ed4-4378-b648-e68912fa54c1"] = "anik@astutemyndz.com"
	userIDEmailMap["ff75bbfb-1f1e-4624-85c8-1e22cfad7021"] = "anuja@auraeco.com"
	userIDEmailMap["dfe32698-0099-4176-8250-0e97b7381260"] = "aoraka@beaconhillsmile.com"
	userIDEmailMap["e2e0f40b-0d51-4658-aedb-22dff43f08f5"] = "arif.essa@transnet.net"
	userIDEmailMap["96bee90c-a2e3-4474-b9e4-e529be5da07b"] = "arif.m@elba.com.sa"
	userIDEmailMap["c07abe29-7a37-4c0b-8891-33465ffc0a22"] = "ashique.iqbal@alfahim.ae"
	userIDEmailMap["406ba35d-a3e3-43fd-8701-6adb4ef34452"] = "ashish.kumar@unibicfoods.in"
	userIDEmailMap["3921cd9c-2cb5-4314-a054-e37cdd3219d3"] = "ashishashuverma222@html.com"
	userIDEmailMap["fb7c127b-16e3-478f-8720-ad860fdcaec3"] = "ashok.rongali@datalake-solutions.com"
	userIDEmailMap["f3f161e5-c665-45e4-b872-dc30c3a6d927"] = "b.ishaq@mazaj.app"
	userIDEmailMap["405ed571-0052-454e-95c7-6be7649ba0db"] = "bala@fragua.co.in"
	userIDEmailMap["3a826b2d-4d3a-4585-aa3c-4de7ae055c8b"] = "beautyworld@bodyfittraining.com"
	userIDEmailMap["e77bd0ba-eaff-4eb4-aa54-83d6370d14d4"] = "binzin.dombin@lafarge.com"
	userIDEmailMap["9a4fe132-af83-4712-8d40-27198895ee95"] = "bongane@galela.com"
	userIDEmailMap["43ee9822-fc6a-4b5f-ba42-69541ad2ca75"] = "CAFE@VRUNDAVANNURSERY.COM"
	userIDEmailMap["eccf08d5-bd3e-47bc-ad3a-9588795df741"] = "ceo@tonsetech.ae"
	userIDEmailMap["805d233f-1af9-4508-afd9-1fd21bb546d4"] = "chibuezeanthonydavid@gamile.com"
	userIDEmailMap["6758576b-6b4d-4767-9739-92e5fb4d0b51"] = "chris@w1music.com"
	userIDEmailMap["83c0739b-d53b-42ca-9473-a58a1a1360d2"] = "clayne.okallo@healthierkenya.co.ke"
	userIDEmailMap["7c5145fb-3c98-4f4d-9f21-144974bf72c9"] = "cleon@magsol.net"
	userIDEmailMap["435eb25a-e176-44b2-a35c-9c7e654e9917"] = "contact@indiabizzness.online"
	userIDEmailMap["505fb87c-1373-404b-8243-05304bea8643"] = "contact@mentorbot.in"
	userIDEmailMap["71aef736-53fc-4887-a7e8-cb697642b1e1"] = "contact@profyogendra.com"
	userIDEmailMap["132bb926-d470-47cb-a3fa-64b7c004552c"] = "daniel.fidelis@venturegardengroup.com"
	userIDEmailMap["89a4b7ef-7c9b-428f-9987-d8735213080b"] = "deepak.ij@yellow.ai"
	userIDEmailMap["1c7e998a-7927-4295-af94-f4ccfbff8135"] = "dhom102040@business.sa"
	userIDEmailMap["d9ab47e0-8894-47c8-b22d-b8d9c31d383b"] = "director@easyloan.ae"
	userIDEmailMap["6ae5d580-5d48-4f6c-88c8-33b887b76af6"] = "Director@shasha.co.zw"
	userIDEmailMap["9920e2df-530b-4126-81cb-c1532f2d617d"] = "dm@panbergco.com"
	userIDEmailMap["9044a44b-6ea1-4192-b243-e0b62e503289"] = "educator@cikguhairul.com"
	userIDEmailMap["31ae6828-984a-428e-b373-eea0d0586442"] = "edwin@avataragency.co.za"
	userIDEmailMap["12eeac13-39a4-4507-99de-146c62ae3757"] = "elmagico124@startpoint.com"
	userIDEmailMap["5bdc2d48-2c8b-4953-8768-0e34bbd09980"] = "erp@daralhay.ae"
	userIDEmailMap["fa9974dc-8fa9-4ac5-9baa-7fa8bdc0d52f"] = "etang.egbe@mtn.com"
	userIDEmailMap["f521b171-8de6-4d58-a40b-4fb414aba848"] = "fatima@matchandstyle.com"
	userIDEmailMap["cd55ac7d-399d-4a28-949a-7f1e1fd40f41"] = "fatimashatimakuburi@gimail.com"
	userIDEmailMap["1eb8f1f1-be3e-4644-8fa3-a869fad1db4f"] = "george@veriphy.co"
	userIDEmailMap["8d952b1f-18c4-4b83-ac70-653a6f62ac02"] = "hamza@arbantech.net"
	userIDEmailMap["40cd5496-bc02-4760-b664-54406b67b2c7"] = "hello@ichhori.com"
	userIDEmailMap["7e6a8c03-2fa5-4e9a-aca9-db7c948103eb"] = "help@urbancompany.com"
	userIDEmailMap["df36061d-93c7-4289-8e48-73ad1a797059"] = "hkgan@hostahome.my"
	userIDEmailMap["f944f5c6-e0f4-425c-b7bc-a8604cb3db97"] = "hussam@keytime.sa"
	userIDEmailMap["6ee643c7-8efa-4e20-b30b-980fde814bb8"] = "ibrahim.mc@mohr.gov.my"
	userIDEmailMap["3cdebdec-45f9-4c0c-9785-643746f3c54a"] = "iconspoint@iconsdcservices.com"
	userIDEmailMap["99eb82b9-d1d2-42e5-be4e-070b7896901d"] = "ildperak@uitm.edu.my"
	userIDEmailMap["780ef59a-7188-420c-84e1-44713a0f7d62"] = "info@bolexvtu.com.ng"
	userIDEmailMap["c09f96d8-ba18-4543-98c0-ff2cf83f5b85"] = "info@buysalewala.com"
	userIDEmailMap["dc4ae0ef-451c-4571-93d2-b6e2f4d7ad6a"] = "info@darum.com.ng"
	userIDEmailMap["8fe009ba-8f8f-451e-a14a-148821e002cf"] = "info@fabugit.com"
	userIDEmailMap["eea99423-4a51-4c9b-8f01-9e2d2b1708af"] = "info@fcl.sa"
	userIDEmailMap["9afaee73-7608-4938-86f6-2291abb82f6c"] = "info@iklinikstores.com"
	userIDEmailMap["d2480f5a-bfdc-4907-aa78-ab62771d0591"] = "info@littletagsluxury.com"
	userIDEmailMap["66c531bd-cbfb-42e7-9c94-e83e6f204d08"] = "info@masternettechnologies.com"
	userIDEmailMap["016a51e1-6816-41a7-879d-f05be59be3ac"] = "info@pmczambia.agency"
	userIDEmailMap["f4775342-d90d-441f-812b-68d9206d8eb4"] = "info@sstinfosolution.in"
	userIDEmailMap["4a24719c-9bb3-42c2-b4b4-8d55da2e9f3b"] = "info@sundaydigitals.com.ng"
	userIDEmailMap["30a273bf-9004-438c-962b-c4cd3d42ab7b"] = "infor@promiseohaneje.com"
	userIDEmailMap["1c0bf359-c1d2-4999-ae62-d12030ce2829"] = "invest@texvalley.info"
	userIDEmailMap["52e75ec5-6300-411d-a5ae-a74142a4ec4c"] = "irsyad@fip.unp.ac.id"
	userIDEmailMap["043df2ee-8aea-4103-b954-8eec5723460a"] = "james@ec.co.zw"
	userIDEmailMap["6b3758ce-82cd-4cc3-a625-82dbb1b25677"] = "jani.pretorius@clickatell.com"
	userIDEmailMap["ca68c310-05e7-4029-b0fe-9ad355c7f02b"] = "jazz@pixod.com"
	userIDEmailMap["eabe372d-2526-4043-afb2-2b30bea44e6d"] = "Jihad@skydreamst.com"
	userIDEmailMap["ab69d8ea-9c5f-44c1-b904-29b20bf6ec8d"] = "jkilusu@barrick.com"
	userIDEmailMap["385df87a-fa05-4be4-9c26-93fda8ff0802"] = "jleon@groundbooker.com"
	userIDEmailMap["ef401e6c-6c7e-43f4-92ec-c45aea9f5233"] = "jleon@groundbooker.com"
	userIDEmailMap["7353a54c-c1cf-454d-96b3-f62631a37a44"] = "jleon@groundbooker.com"
	userIDEmailMap["a00e56a1-d065-4e2c-b5fb-a29e2ba6e2b5"] = "julien@paneo.tech"
	userIDEmailMap["9e3f10c2-4964-4ab9-837a-c6ea93b2c9bc"] = "justusomare@9999gmail.com"
	userIDEmailMap["4cf26995-fe3c-4507-a38a-b64ea6031f70"] = "kabirpatil@clinitechlab.com"
	userIDEmailMap["3f28c55f-7cf4-40ce-ad12-896fcc4b502a"] = "kapil@bankify.money"
	userIDEmailMap["eef61511-2e61-4dcd-888a-43d42e0a604a"] = "lindokuhle@msdfgroup.co.za"
	userIDEmailMap["82087961-a30f-4f6b-b0b1-6196eb695f4d"] = "Mahmoud.H@fnatek.ae"
	userIDEmailMap["b7d4c51f-05e8-465d-80ba-68b7b339fb83"] = "marcoms01.sg@cornerstonewines.com"
	userIDEmailMap["3baa25f2-6b8d-4ee5-be65-5d3d4250405a"] = "marketing@banke.ae"
	userIDEmailMap["56cddc2b-b44b-40ed-87d2-f2d087b50898"] = "marketing@sakuraacafe.com"
	userIDEmailMap["8b37d2e7-61d1-4af8-b266-9c1044dc46c1"] = "marketing@scarletn.com"
	userIDEmailMap["4000e364-f522-4efc-b11a-30acf8fb224f"] = "masoodahmed@transpost.co"
	userIDEmailMap["378cd974-f847-45b0-8c93-c2759994734e"] = "mayank@kam.ae"
	userIDEmailMap["d60233f5-1608-4cb6-8509-f43ddbdd351a"] = "medhat@alrayyan.site"
	userIDEmailMap["e17ff610-dea8-4c00-acb6-e9176161c35d"] = "merabaenock@united.co.ke"
	userIDEmailMap["9cbd7a60-1f80-44af-b45b-a4105dde8647"] = "mohamed.alshafea@ese.gov.ae"
	userIDEmailMap["08aa7ef1-8639-4970-8bd4-5c4a663af0d1"] = "ms@algebracorp.com"
	userIDEmailMap["52eee3b0-310f-493f-b755-6ee4e72248fb"] = "mundzir@utmholdings.com"
	userIDEmailMap["19851b90-0407-4c79-974c-046da1fe7144"] = "narayan@goodseeds.in"
	userIDEmailMap["b76819dc-fb43-42ca-8597-681e55fe334b"] = "neeraj70@gimail.com"
	userIDEmailMap["4d08ac40-1b3d-4e10-9525-6c109bef5972"] = "nimrah@kwsme.com"
	userIDEmailMap["81197696-5d53-48c1-8850-ed5f11070a00"] = "noemi@metaphysicalanatomy.com"
	userIDEmailMap["9f0d8730-d563-4bbf-9218-82b6d595aabf"] = "norazwankhairi@nakmnr.onmicrosoft.com"
	userIDEmailMap["a780902c-39ce-4680-8ab7-2867583b7ee1"] = "oatulomah@aust.edu.ng"
	userIDEmailMap["0afb71bb-f610-4158-8a8c-55eaf5be03c6"] = "oluabusomma.kizor-akaraiwe@stu.cu.edu.ng"
	userIDEmailMap["aaedc9e8-5f37-48f1-8708-541241eb8770"] = "olumide@smartedgebusiness.com"
	userIDEmailMap["364fee9b-fdc2-4730-b707-7e08fd137ca3"] = "p.silas@havillauniversity.edu.ng"
	userIDEmailMap["9401d319-f34c-41ca-8adc-d391e8b01c94"] = "partnerships@blenditrawapothecary.in"
	userIDEmailMap["c2e6cc23-b0f3-4686-bbf5-6dd06c34f529"] = "pelumi@solidfiction.com"
	userIDEmailMap["5a39b5bf-6710-402a-94a6-6475f440c727"] = "prince@xtremerepublic.co.ke"
	userIDEmailMap["598acb29-33ca-40d0-8915-07fd9d940da6"] = "priya@verofax.com"
	userIDEmailMap["021bfa0a-dae5-4698-be4b-2fd7652a84e5"] = "priyanka@carerforcancer.com"
	userIDEmailMap["9a7043ee-dfc0-4fb6-a64d-86b2c4e8be35"] = "rajnish.verma@rkautohouse.com"
	userIDEmailMap["db1103ac-3541-49bd-810a-12b707cf174c"] = "ramdhan@amanahcorp.co.id"
	userIDEmailMap["600b496c-4273-4592-81e1-85f69ce50e9a"] = "ramesh.varadharajan@jbmgroup.com"
	userIDEmailMap["3658dd17-5a7c-43a4-8d0a-69d541299f36"] = "ramy.shahin@mobilitycairo.com"
	userIDEmailMap["6b39882b-5f6d-4b79-b653-4148c79063a0"] = "reshma@semplinet.net"
	userIDEmailMap["3af49cd5-36c7-465b-bfda-11d1210d7d78"] = "richard.konlan@uicghana.org"
	userIDEmailMap["eb704381-ec62-4f3e-9928-3d2c75915e72"] = "rujendra.maharjan@tai.com.np"
	userIDEmailMap["f37ace58-8dea-4b3d-9039-ec65c123340b"] = "s.abukarar@transportation.gov.org"
	userIDEmailMap["2871a986-4af2-4078-b866-41b0ebafd969"] = "s.musa@alhayah.edu.sa"
	userIDEmailMap["6f84136d-12cf-4392-9b1f-5738e7bb85a6"] = "samah@keepyousocial.com"
	userIDEmailMap["88ae2aa2-ec73-4164-9118-f7a10966b10b"] = "sara@elbeih.com"
	userIDEmailMap["3ff04ec5-7fb5-48c9-8c3f-f953155fe97e"] = "shivaabhinay.bandaru@mindgraph.com"
	userIDEmailMap["5f2536cb-8f4f-4d0f-bf36-6763824d6638"] = "shivam@yellow.ai"
	userIDEmailMap["bfe8eb85-b085-41ab-8b5e-053bd27cf5e1"] = "shivamtr@yell.ai"
	userIDEmailMap["5f2536cb-8f4f-4d0f-bf36-6763824d6638"] = "shivamtrt@yell.com"
	userIDEmailMap["5f2536cb-8f4f-4d0f-bf36-6763824d6638"] = "shivamtuy@yell.ai"
	userIDEmailMap["e84f68bc-330f-45de-8739-a14e482a4583"] = "shiyas.khan@autodieselpartsgroup.com"
	userIDEmailMap["dbb876cc-4067-4b32-b66c-cbd554f64b46"] = "sjupe2111@ueab.ac.ke"
	userIDEmailMap["d2c47503-e086-4e8a-8b25-628ca896fdb5"] = "sona.s@elabram.com"
	userIDEmailMap["3aec7544-57e5-4f18-a0fa-2195b8274884"] = "support@easygas.my"
	userIDEmailMap["c2fddcda-6cb4-4e89-b394-27977ffe61e5"] = "support@gmobile.com.ng"
	userIDEmailMap["659539e0-ca5f-473f-be8c-2f26878b1881"] = "support@hnk9.in"
	userIDEmailMap["1743af90-fdcc-4d4e-a784-77f1cb9689ee"] = "support@kfexchange.com"
	userIDEmailMap["563da462-4b20-4c60-909d-57cc9a92aadc"] = "tanay.prajapati@relevel.com"
	userIDEmailMap["32a5e356-ff41-4463-bf0f-ccd11db8af6b"] = "tannu@getyouat.com"
	userIDEmailMap["2212db8f-29f5-4242-ab02-62faed27690f"] = "tawfigmadni123@gampl.com"
	userIDEmailMap["897e3e88-c73e-4a16-98dd-9a8fb877ebe2"] = "test@google.com"
	userIDEmailMap["897e3e88-c73e-4a16-98dd-9a8fb877ebe2"] = "test@google.com"
	userIDEmailMap["dc513571-12dc-4c43-a93e-f9733eee8c47"] = "test@jmail.com"
	userIDEmailMap["dc513571-12dc-4c43-a93e-f9733eee8c47"] = "test@kmail.com"
	userIDEmailMap["117ae3b3-4efb-48e0-86ba-58a364699d12"] = "tg@figtree.com.sg"
	userIDEmailMap["3cbe3ec8-6aee-4f89-85cb-73619492dcaa"] = "thileep@uniex.in"
	userIDEmailMap["e3a45556-7955-4fd3-9f4f-d9cd1d0e6eb1"] = "tim@smashverse.com"
	userIDEmailMap["4be27cf0-6768-4dd9-ae3e-df288182a1be"] = "tmaleswena@sword-sa.com"
	userIDEmailMap["f04f4726-423e-4c6f-98bc-9495e1bd3e93"] = "wjeezAdmin@wjeez.onmicrosoft.com"
	userIDEmailMap["4bdd3cba-bbf1-49b2-a747-956ac7124ecd"] = "work@afrogazette.co.zw"
	userIDEmailMap["a6c5bea3-f124-43af-b8db-a2664408a016"] = "xebag59235@hauzgo.com"
	userIDEmailMap["f0bc2d70-d5ef-4b0f-8f07-36e0a3be6d07"] = "yadwinder.adcw@adcwsangrur.com"
	userIDEmailMap["b523b29e-9131-40ce-bea9-9d6b2a71487e"] = "yousef@allam.solutions"
	return userIDEmailMap
}
