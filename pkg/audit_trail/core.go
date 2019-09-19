package audit_trail

import (
	"fmt"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v3"
)

type ICore interface {
	Insert(tx *sqlx.Tx, data_audit *AuditTrail) (err error)
}

// core contains db client
type core struct {
	db    *sqlx.DB
	redis *redis.Pool
	pid   int64
}

const redisPrefix = "molanobar-v1"

func (c *core) Insert(tx *sqlx.Tx, data_audit *AuditTrail) (err error) {
	data_audit.Timestamp = time.Now()
	query := `INSERT INTO mla_logs(user_id, query_executed, table_name, project_id,timestamp) VALUES (?,?,?,?,?)`
	_, err = tx.Exec(query, data_audit.UserID, data_audit.Query, data_audit.TableName, c.pid, data_audit.Timestamp)
	if err != nil {
		return err
	}
	return err
}

func ConstructLogQuery(query string, args ...interface{}) string {
	getParamString := func(v interface{}) string {
		switch v.(type) {
		case int:
			return fmt.Sprintf("%d", v.(int))
		case int16:
			return fmt.Sprintf("%d", v.(int16))
		case int64:
			return fmt.Sprintf("%d", v.(int64))
		case float64:
			return fmt.Sprintf("%f", v.(float64))
		case string:
			return fmt.Sprintf("'%s'", v.(string))
		case time.Time:
			return fmt.Sprintf("'%s'", v.(time.Time))
		case null.Time:
			return fmt.Sprintf("'%v'", v.(null.Time))
		case bool:
			if v.(bool) {
				return "1"
			}
			return "0"
		default:
			return "'unsupported type'"
		}
	}

	newQuery := query
	for _, v := range args {
		newQuery = strings.Replace(newQuery, "?", getParamString(v), 1)
	}
	return newQuery
}
