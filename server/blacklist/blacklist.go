package blacklist

import (
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

const batchSize = 999 // maximum variable value for sqlite3

type Loader struct {
	urls         []string
	repo         Repository
	syncInterval time.Duration
}

func NewLoader(urls []string, repo Repository, syncInterval time.Duration) Loader {
	return Loader{urls, repo, syncInterval}
}

func (l Loader) StartSync() {
	for {
		startTime := time.Now()

		l.load(startTime)

		endTime := time.Now()
		timeTaken := time.Since(startTime) - time.Since(endTime)
		if timeTaken < l.syncInterval {
			time.Sleep(l.syncInterval - timeTaken)
		} else {
			log.Warn("Synchronization took longer than specified syncInterval, skipping sleep")
		}
	}
}

func (l Loader) load(updateTimestamp time.Time) {
	log.Info("Started blacklist synchronization")
	nextIndex := 0
	hostsToInsert := make([]string, batchSize)
	for _, url := range l.urls {

		resp, err := http.Get(url)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "url": url}).Error("Coudn't get black list from url")
			continue
		}
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			hostsToInsert[nextIndex%batchSize] = scanner.Text()
			nextIndex++
			if nextIndex%batchSize == 0 {
				l.repo.addBatchToDB(hostsToInsert, updateTimestamp)
				hostsToInsert = make([]string, batchSize)
			}
		}
	}
	if nextIndex%batchSize != 0 {
		l.repo.addBatchToDB(hostsToInsert[0:nextIndex%batchSize], updateTimestamp)
	}
	l.repo.deleteEntriesBefore(updateTimestamp)
	log.WithField("fetched hosts", nextIndex).Info("Finished blacklist synchronization")

}

type Repository interface {
	deleteEntriesBefore(limit time.Time)
	addBatchToDB(hosts []string, timestamp time.Time)
}

type SqlRepository struct {
	db *sql.DB
}

func NewSqliteRepository(db *sql.DB) SqlRepository {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS blacklist (host varchar(255) PRIMARY KEY, ts timestamp )")
	if err != nil {
		log.WithField("error", err).Error("Couldn't create blacklist table in database")
	}
	return SqlRepository{db}
}

func (r SqlRepository) deleteEntriesBefore(limit time.Time) {
	_, err := r.db.Exec("DELETE from blacklist where ts < ?", limit.Format(time.RFC3339))
	if err != nil {
		log.WithField("error", err).Error("Couldn't delete blacklist entries")
	}
}

func (r SqlRepository) addBatchToDB(hosts []string, timestamp time.Time) {
	placeholder := fmt.Sprintf("(?,'%s')", timestamp.Format(time.RFC3339))

	hostParams := make([]interface{}, len(hosts))
	for i, e := range hosts {
		hostParams[i] = e
	}

	stmt := fmt.Sprintf("REPLACE INTO blacklist (host, ts) VALUES %s%s", strings.Repeat(placeholder+", ", len(hosts)-1), placeholder)

	_, err := r.db.Exec(stmt, hostParams...)
	if err != nil {
		log.WithField("error", err).Error("Couldn't insert blacklist entries")
	}
}

func (r SqlRepository) IsBlacklisted(host string) bool {
	var found bool
	err := r.db.QueryRow("SELECT EXISTS (SELECT 1 FROM blacklist WHERE host=?)", host).Scan(&found)
	if err != nil {
		log.WithField("error", err).Error("Couldn't query blacklist")
	}
	return found
}
