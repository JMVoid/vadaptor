package mysql

import (
	"testing"
)


const DBCONFIG = "root:root@tcp(127.0.0.1:3306)/ssrpanel"


//func TestPullUser(t *testing.T){
//
//	client := NewDb(DBCONFIG)
//	ur, err:= client.PullUser()
//	if err != nil {
//		t.Error("error on pulluser", err)
//		return
//	}
//
//	if len(ur.Usermap) > 0 {
//		t.Log("fetch user list successfully ")
//		return
//	}
//	t.Error("error on pulluser", err)
//
//}
//
//
//func TestPushUserTransfer(t *testing.T){
//
//	userRepo := new(pb.UserRepo)
//	user1 := new(pb.User)
//	user1.Username = "hkdollar@outlook.com"
//	user1.UpIncr = 18200
//	user1.DownIncr = 21000
//
//	userRepo.Usermap = make(map[string]*pb.User)
//	userRepo.Usermap["hkdollar@outlook.com"] = user1
//
//	client := NewDb(DBCONFIG)
//	err := client.PushUserTransfer(userRepo)
//	if err != nil {
//		t.Error("error on push user transfer", err)
//	}
//
//	t.Log("check the data in db or not")
//}


func TestPushNodeStatus(t *testing.T) {

	client := NewDb(DBCONFIG)
	err := client.PushNodeStatus(2, 888888, "0.85,0.85,0.85")
	if err != nil {
		t.Error("fail to push node status")
		return
	}
	t.Log("push node status db action is work")
}