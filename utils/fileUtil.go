package utils

import (
	"github.com/JMVoid/v2ssadaptor/pb"
	"github.com/gogo/protobuf/proto"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os/exec"
	"bytes"
)

const DefaultFileMode = 0644

func WriteRepo(fname string, userRepo *pb.UserRepo) error {


	data, err := proto.Marshal(userRepo)
	if err != nil {
		log.Errorf("fail to marshal UserRepo object with %v\n", err)
		return err
	}

	if err = ioutil.WriteFile(fname, data, DefaultFileMode); err != nil {
		log.Errorln(err)
		return err
	}

	return nil
}


func ReadRepo(fname string) (userRepo *pb.UserRepo, err error){


	 data, err := ioutil.ReadFile(fname)
	 if err != nil {
		log.Errorln(err)
		return nil, err
	 }

	 ur := new(pb.UserRepo)
	 if err = proto.Unmarshal(data, ur); err != nil {
	 	log.Errorf("error on unmarshal: %v\n", err)
		return nil, err
	 }

	return ur, nil
}


func GetLoadAvg() string{
	var out bytes.Buffer
	loadStr := "cat /proc/loadavg | awk '{print $1\" \"$2\" \"$3}'"
	cmd := exec.Command("bash", "-c", loadStr)
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Error(err)
		return ""
	}
	return out.String()
}