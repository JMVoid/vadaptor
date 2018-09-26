package controller

import (
	"v2ray.com/core/app/proxyman/command"
	statscmd "v2ray.com/core/app/stats/command"
	"context"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/common/protocol"
	"v2ray.com/core/proxy/vmess"
	log "github.com/Sirupsen/logrus"
	"fmt"
	"github.com/JMVoid/v2ssadaptor/pb"
	"time"
	"google.golang.org/grpc"
)

type V2Controller struct {
	client      command.HandlerServiceClient
	statsClient statscmd.StatsServiceClient

	inBoundTag string
}

const (
	UplinkFormat   = "user>>>%s>>>traffic>>>uplink"
	DownlinkFormat = "user>>>%s>>>traffic>>>downlink"
)

func NewV2Controller(addr, tag string) (*V2Controller, error) {

		cc, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil  {
			log.Error(err)
			return nil, err
		}

	client := command.NewHandlerServiceClient(cc)
	statsClient := statscmd.NewStatsServiceClient(cc)

	vc := &V2Controller{
		client:      client,
		statsClient: statsClient,
		inBoundTag:  tag,
	}
	return vc, nil

}

func (v *V2Controller) AddUser(u pb.User) error {

	var err error
	for count :=0; count<10; count++{
		_, err = v.client.AlterInbound(context.Background(), &command.AlterInboundRequest{
			Tag: v.inBoundTag,
			Operation: serial.ToTypedMessage(
				&command.AddUserOperation{
					User: &protocol.User{
						Level: u.GetLevel(),
						Email: u.GetUsername(),
						Account: serial.ToTypedMessage(&vmess.Account{
							Id:               u.GetUuid(),
							AlterId:          u.GetAlterId(),
							SecuritySettings: &protocol.SecurityConfig{Type: protocol.SecurityType_AUTO},
						}),
					},
				}),
		})

		if err == nil {
			//log.Errorf("failed to call add user: %v\n", err)
			break
		}
		time.Sleep(time.Duration(200) * time.Millisecond)
	}
	if err != nil {
		return err
	}

	log.Printf("add [%s] user to v2ray successfully", u.GetUsername())
	return nil
}

func (v *V2Controller) RemoveUser(u pb.User) error {
	resp, err := v.client.AlterInbound(context.Background(), &command.AlterInboundRequest{
		Tag: v.inBoundTag,
		Operation: serial.ToTypedMessage(
			&command.RemoveUserOperation{
				Email: u.GetUsername(),
			}),
	})
	if err != nil {
		//log.Errorf("failed to call remove user : %v", err)
		return err
	}
	log.Printf("remove user [%s] successfully with %v", u.GetUsername(), resp)
	return nil
}

func (v *V2Controller) GetTraffic(u *pb.User, isReset bool) error {

	ctx := context.Background()
	up, err := v.statsClient.GetStats(ctx, &statscmd.GetStatsRequest{
		Name:   fmt.Sprintf(UplinkFormat, u.GetUsername()),
		Reset_: isReset,
	})

	if err != nil {
		log.Errorf("get upload traffic user %s error %v", u.GetUsername(), err)
		return err
	}

	down, err := v.statsClient.GetStats(ctx, &statscmd.GetStatsRequest{
		Name:   fmt.Sprintf(DownlinkFormat, u.GetUsername()),
		Reset_: isReset,
	})

	if err != nil {
		log.Errorf("get download traffic user %s error %v", u.GetUsername(), err)
		return err
	}

	u.UpIncr += up.Stat.Value
	u.DownIncr += down.Stat.Value
	return nil

}
