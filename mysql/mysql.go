package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/Sirupsen/logrus"
	"github.com/JMVoid/vadaptor/pb"
	"github.com/JMVoid/vadaptor/utils"
	"crypto/tls"
	"github.com/go-sql-driver/mysql"
	"io/ioutil"
	"crypto/x509"
)

type DbClient struct {
	SrvCfg string
}

//func NewDb(dbCfg string) *DbClient {
func NewDb(dbCfg utils.V2ray) *DbClient {

	var certs tls.Certificate
	var err error
	var rootCertPool *x509.CertPool

	if dbCfg.DbSslCa != "" {
		rootCertPool = x509.NewCertPool()
		pem, err := ioutil.ReadFile(dbCfg.DbSslCa)
		if err != nil {
			log.Panicln(err)
		}
		if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
			log.Panicln("Failed to append PEM.")
		}
	}

	if dbCfg.DbSslCert != "" && dbCfg.DbSslKey != "" {
		certs, err = tls.LoadX509KeyPair(dbCfg.DbSslCert, dbCfg.DbSslKey)
		if err != nil {
			log.Panicln("config SSL but fail to load cert and key pem files")
		}
	}

	clientCert := make([]tls.Certificate, 0, 1)
	clientCert = append(clientCert, certs)

	mysql.RegisterTLSConfig("mysqlSSL", &tls.Config{
		RootCAs:      nil,
		Certificates: clientCert,
	})

	dbClient := new(DbClient)
	dbClient.SrvCfg = dbCfg.DbCfg + "?tls=mysqlSSL&tls=skip-verify"
	return dbClient
}

func (s *DbClient) PullUser(nodeId uint32) (userRepo *pb.UserRepo, err error) {

	db, err := sql.Open("mysql", s.SrvCfg)

	if err != nil {
		log.Errorf("fail to open mysql server with : %v", err)
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query("SELECT username, vuuid, valterid, vlevel, enable, transfer_enable, u, d FROM ssrpanel.user WHERE enable > 0 "+
		"AND STATUS IN (0,1) AND id in "+
		"(SELECT ul.user_id FROM ss_node_label snl, user_label ul "+
		"WHERE ul.label_id = snl.label_id "+
		"AND snl.node_id = ?)", nodeId)
	if err != nil {
		//log.Errorf("fail to pull user from db with err: %v", err)
		return nil, err
	}
	defer rows.Close()

	userRepo = new(pb.UserRepo)
	userRepo.Usermap = make(map[string]*pb.User)
	for rows.Next() {
		user := new(pb.User)
		rows.Scan(&user.Username, &user.Uuid, &user.AlterId, &user.Level, &user.Enable, &user.TransferEnable, &user.Uplink, &user.Downlink)
		log.Debugf("%s-%s-%d-%d-%d-%d-%d-%d\n", user.Username, user.Uuid, user.AlterId, user.Level, user.Enable, user.TransferEnable, user.Uplink, user.Downlink)

		// should check the email is double or not
		userRepo.Usermap[user.Username] = user

	}
	return userRepo, nil
}

func (s *DbClient) PushUserTransfer(userRepo *pb.UserRepo) error {
	var queryWhend string
	var queryWhenu string
	var sqlWhen string
	var sqlWhere string
	var wSql string

	var userList []pb.User

	uLen := len(userRepo.Usermap)
	if uLen < 1 {
		return nil
	}

	for _, v := range userRepo.Usermap {
		userList = append(userList, *v)
	}

	queryHeader := "update user set "
	for i := 0; i < uLen; i++ {
		queryWhenu += fmt.Sprintf("WHEN '%s' THEN u + %d ", userList[i].Username, userList[i].UpIncr)
		queryWhend += fmt.Sprintf("WHEN '%s' THEN d + %d ", userList[i].Username, userList[i].DownIncr)
	}

	for j := 0; j < uLen; j++ {
		wSql += fmt.Sprintf("'%s'", userList[j].Username)
		if j >= 0 && j < uLen-1 {
			wSql += ", "
		}
	}
	sqlWhen = fmt.Sprintf("u = CASE username %s END, d = CASE username %s END", queryWhenu, queryWhend)
	sqlWhere = fmt.Sprintf(" WHERE username IN (%s)", wSql)

	sqlRun := queryHeader + sqlWhen + sqlWhere

	db, err := sql.Open("mysql", s.SrvCfg)
	if err != nil {
		log.Errorf("fail to open mysql server with : %v", err)
		return err
	}

	defer db.Close()

	log.Debugln(sqlRun)
	stmt, err := db.Prepare(sqlRun)
	if err != nil {
		log.Errorf("Push user prepare error:%v", err)
		return err
	}

	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		log.Error(err)
	}
	rows.Close()

	return nil
}

func (s *DbClient) PushNodeStatus(nodeId uint32, upTime int64, load string) error {
	queryHeader := "INSERT INTO ss_node_info(`id`, `node_id`, `uptime`, `load`, `log_time`) values "
	valueStr := fmt.Sprintf("(NULL, %d, %d, '%s', unix_timestamp())", nodeId, upTime, load)

	sqlRun := queryHeader + valueStr

	db, err := sql.Open("mysql", s.SrvCfg)
	if err != nil {
		log.Errorf("fail to open mysql server with : %v", err)
		return err
	}

	defer db.Close()
	stmt, err := db.Prepare(sqlRun)

	if err != nil {
		log.Errorf("Push Node status prepare error:%v", err)
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	defer rows.Close()
	if err != nil {
		log.Error(err)
	}
	return nil
}

//create sample table
// email string, uuid string, alerId int, level int, u bigint, d bigint

// create sql string to read all user from mysql

// create sql update with case when grammar to add transfer increment
