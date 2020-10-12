# datagen
Data Generation Tool

Import V1 branch 

Navigate into datagen folder

RUN export PATH=$PATH:/usr/local/go/bin [To set go variables. Skip it if already set]

go get gopkg.in/yaml.v2
go get gopkg.in/natefinch/lumberjack.v2

There will be 'output.txt' int /datagen. you can either flush the contents/delete the file to have just the output of this run. Or change the variable: output_file_name in file: sampleconfigV2.yaml

RUN go build main.go

RUN go run main.go

DataGen Code - Logic: 
1. Create Output directory locally - output_data
2. All files will have the current date/time in file format - debuglogs, errorlogs, outputdata, userdata
3. Register (Instantiate) "LogWriter" - Console will have nothing except the log file names for that run
4. Read the input config file and create the "configuration" object which is used for all the further operations
5. Instantiate "OutputFileWriter" and "UserDataWriter"
6. Generate Data


Generate Data Logic
1. Read "UserData_*" files from cloud storage and Process the files to read and store all the existing user data to a map (existingUsers)
2. Config structure looks like the following

ListOfGlobalAttributes - TotalExecutionTime
ListofSegments(noOfUsersInSegment, on/offSwitchForAttributesInOutput, DefaultTickerTimeBetweenEvents(can be overidden at transition level),  SegmentConfiguration, SeedTime)
SegmentConfiguration
{
	// all of probablities should be 1. Check enforced in code 
	ActivityProbablityMap - Probablity of someone (doing something in the system, doing nothing, exit)
	EventProbablityMap - Event Correlation Matrix, Seed_events (What are possible start events for the whole flow Eg: home page, some campaign view), Exit events (Eg: Payment submit, Logout)
	EventAttributes - Predefined, Default, Custom
	UserAttributes - Default, custom
	EventDecorators
	OverrideRules
}
Predefined attributes - These attributes must be in output and the value should be from one in the list. This is declared one per event Eg: Event E1 will have attribute A1 and its value will be one among (a1,a2,a3) but Event E2 will have attribute A2 with value one among (a4,a5)
Default attributes - These attributes must be in output and the value should be from one in the list. This is common across event/users
Eg: Age (18-25, >25, <18)
Custom attributes - These are like default attributes but they might/might not be present in the output

All the attributes are predefined once before the run and retained through out the run (For both users and events)

Event decorators - Eg: there is a design listing page and the user clicks these designs one by one and views them. In this case, the page is same but the designs he is viewing is different. We cant use attribute here since it is initiated once. Decorators share the same structure as attributes but are picked during execution

Override Rules : If certain default probablities/tickertime need to be modified based on presence of certain attributes

Event Correlation matrix: Events and transition probablities between events

3. Compute Range Maps for Each probablity map 
Eg : A1{a1: 0.2, a2: 0.3, a3: 0.5} RangeMap = (0-1)a1, (2-4)a2, (5-9)a3 Multiplier-10

4. Spawn two threads 
	- Operate on each segment (Run segments in parellel)
	- Insert new users to system on any available segment based of new_user_seed_duration/no_of_new_users_seed_count 

5. Segment Operation
	- Run for all the users in parallel - pick a new user/existing user again based on some probablity - if its a new user, add the user to user data file and pick attributes from user attributes map
	- Check for a random activity
	- If activity = DoSomething
		Get Random event (pick from seed if its first/ pick from the correlation map depending on the last known state)
		if there exist an overridden rule for the transition, then pick the transition apply the rule, recompute the probablity(Relatively assign/adjust other event's probablity) and pick one event after recomputation
		Pick event attributes
	-set event attributes
	- set event decorators
	- add the output to file

6. Upload files to cloud - logs/output/userdata
