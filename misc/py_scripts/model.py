import pandas as pd
import json 
import os

class Model():

    def __init__(self,path):
        self.path = path
        self.full_paths = []
        self.get_files()


    def get_files(self):
        chunk_files = os.listdir(self.path)
        for f in chunk_files:
            tmp = os.path.join(self.path,f)
            self.full_paths.append(tmp)
        
    
    def read_chunks(self,fn):
        for f in self.full_paths:
            with open(f,"r") as m:
                for line in m:
                    fn(line)

        return None
                    
    def print_events_names(self):
        count=0
        t1={}
        t2={}
        t3={}
        t1[1]=0
        t1[3]=0
        t1[2]=0
        t1[3]=0
        t2[1]=0
        t2[2]=0
        t2[3]=0
        t2[4]=0
        t3[1]=0
        t3[2]=0
        t3[3]=0
        t3[4]=0
        for f in self.full_paths:
            with open(f,"r") as m:

                for line in m:
                        event = json.loads(line)
                        if len(event['rp']['en'])>0:
                            a=event['rp']['ufp']
                            b=0
                        
                            
                            if event['rp']['ouc']>0:
                                k = "{} -> {} -> {}".format(event['rp']['en'],event['rp']['ouc'],len(a))
                                count+=1
                                print(k)
                                t2[len(event['rp']['en'])]+=1
                            else:
                                t1[len(event['rp']['en'])]+=1
                        t3[len(event['rp']['en'])]+=1
                            
        print(count,t1,t2,t3)

    def search_pattern(self,patt):
        for f in self.full_paths:
            with open(f,"r") as m:
                for line in m:
                    event = json.loads(line)
                    if event['rp']['en']==patt:
                        a=event['rp']['ufp']
                        k = "{} -> {} -> {}".format(event['rp']['en'],event['rp']['ouc'],len(a))
                        print(k)
                        break


    def print_pattern_containing(self,patt):
        for f in self.full_paths:
            with open(f,"r") as m:
                for line in m:
                    event = json.loads(line)
                    if patt in event['rp']['en']:
                         k = "{} -> {}".format(event['rp']['en'],event['rp']['ouc'])
                         print(k)


    def print_pattern_count(self,patt,count):
        for f in self.full_paths:
            with open(f,"r") as m:
                for line in m:
                    event = json.loads(line)
                    if patt in event['rp']['en'] and event['rp']['ouc']>=count :
                         k = "{} -> {}".format(event['rp']['en'],event['rp']['ouc'])
                         print(k)
      
                        
if __name__=="__main__":
    path = "/usr/local/var/factors/cloud_storage/projects/1000002/models/1668307263360/chunks/"
    # path="/usr/local/var/factors/cloud_storage/projects/1000002/models/1668322624695/chunks/"
    path=  "/usr/local/var/factors/cloud_storage/projects/1000002/models/1668579165727/chunks/"
    # path="/Users/vinithkumar/work/scratch/"
    M = Model(path)
    M.print_events_names()
    # M.search_pattern(["$session","$form_submitted"])
    # M.print_pattern_containing('$form_submitted')
    # M.print_pattern_count('$form_submitted',25)