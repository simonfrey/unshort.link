package blacklist

import (
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)


type MockRepo struct {
	deletedCalls []time.Time
	addedHosts   []hostEntry
}

type hostEntry struct {
	host string
	ts   time.Time
}

func (m *MockRepo) deleteEntriesBefore(limit time.Time) {
	m.deletedCalls = append(m.deletedCalls, limit)
}
func (m *MockRepo)  addBatchToDB(hosts []string, ts time.Time) {
	for _, host := range hosts {
		m.addedHosts = append(m.addedHosts, hostEntry{host,ts,})
	}
}


func Test_Loader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, "Hello\nBye")
	}))
	repo := new(MockRepo)
	loader := NewLoader([]string{ts.URL}, repo, time.Hour)
	reloadTs := time.Now()
	loader.load(reloadTs)

	assert.Len(t, repo.deletedCalls, 1)
	assert.Equal(t, repo.deletedCalls[0], reloadTs )

	assert.Len(t, repo.addedHosts, 2)
	assert.Contains(t, repo.addedHosts, hostEntry{"Hello", reloadTs})
	assert.Contains(t, repo.addedHosts, hostEntry{"Bye", reloadTs})
}

func Test_Loader_PeriodicSync(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, "Hello\nBye")
	}))
	repo := new(MockRepo)
	loader := NewLoader([]string{ts.URL}, repo, time.Second)
	tsBefore := time.Now()

	go loader.StartSync()
	time.Sleep(time.Second)

	assert.Len(t, repo.deletedCalls, 1)
	assert.True(t, repo.deletedCalls[0].After(tsBefore))
	assert.True(t, repo.deletedCalls[0].Before(time.Now()))

	assert.Len(t, repo.addedHosts, 2)
}

func Test_SqliteRepo_Add_Host(t *testing.T) {
	sqlite3, _ := sql.Open("sqlite3", ":memory:")
	repository := NewSqliteRepository(sqlite3)

	assert.False(t, repository.IsBlacklisted("Abc"))
	assert.False(t, repository.IsBlacklisted("cde"))

	repository.addBatchToDB([]string{"Abc", "cde"}, time.Now())

	assert.True(t, repository.IsBlacklisted("Abc"))
	assert.True(t, repository.IsBlacklisted("cde"))
}

func Test_SqliteRepo_Remove(t *testing.T) {
	sqlite3, _ := sql.Open("sqlite3", ":memory:")
	repository := NewSqliteRepository(sqlite3)
	insertTs := time.Now()

	repository.addBatchToDB([]string{"Abc", "cde"}, insertTs)

	assert.True(t, repository.IsBlacklisted("Abc"))
	assert.True(t, repository.IsBlacklisted("cde"))

	repository.deleteEntriesBefore(insertTs.Add(time.Hour))

	assert.False(t, repository.IsBlacklisted("Abc"))
	assert.False(t, repository.IsBlacklisted("cde"))

}
