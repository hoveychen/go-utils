package mongo

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"

	"github.com/globalsign/mgo"
	goutils "github.com/hoveychen/go-utils"
	"github.com/hoveychen/go-utils/flags"
	"github.com/pkg/errors"
)

var (
	dbRouterJson = flags.String("dbRouterJson", "", "Router config for different servers. If empty, environment $DB_ROUTER will be adapted.")

	defaultRouter *Router
	lock          sync.Mutex
	once          sync.Once
)

// Router is a smart proxy to open a proper mongodb client determined by the required database
// and/or collection combination.
type Router struct {
	Servers []*Server `json:"servers"`
	Hints   []*Hint   `json:"hints"`

	HintMap map[string]*Server `json:"-"`
}

type Server struct {
	// Name is required to identify the server. MUST be unique in the configuation.
	Name string `json:"name"`
	// Address will be directly passed to mgo.Dial()
	Address string `json:"address"`
}

type Hint struct {
	// ServerName is used to specify the server.
	ServerName string `json:"server_name"`

	// Database and Collection is used to determine the whether this server
	// should be used. It follows the rule to choose the server:
	// 1. Both database and collection argument are perfectly matched.
	// 2. If no such match, the LAST hint matching the db.
	// 3. If no such match(no such database in the config), the LAST hint matching the collection.
	// 4. Choose the first server given in Servers section.
	Database   string `json:"database"`
	Collection string `json:"collection"`
}

func LoadJsonFileRouter(filename string) (*Router, error) {
	d, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "Readfile: %s", filename)
	}
	r := &Router{}
	if err := json.Unmarshal(d, r); err != nil {
		return nil, errors.Wrap(err, "Unmarshall router json")
	}
	r.Init()

	return r, nil
}

// Init must be called before any usage.
// It will perform validation to the router configuration.
func (r *Router) Init() error {
	if err := r.validate(); err != nil {
		return errors.Wrap(err, "Validate")
	}

	serverMap := map[string]*Server{}

	for _, s := range r.Servers {
		serverMap[s.Name] = s
	}
	r.HintMap = map[string]*Server{}
	for _, h := range r.Hints {
		s := serverMap[h.ServerName]
		r.HintMap[h.Database+":"+h.Collection] = s
		r.HintMap["db@"+h.Database] = s
		r.HintMap["c@"+h.Collection] = s
	}

	return nil
}

func (r *Router) validate() error {
	if len(r.Servers) == 0 {
		return errors.New("At least one server should be specified in mongo router config.")
	}

	serverMap := map[string]string{}
	for _, s := range r.Servers {
		if serverMap[s.Name] != "" {
			return errors.Errorf("Server name duplication, %s, %s, %s", s.Name, s.Address, serverMap[s.Name])
		}
		serverMap[s.Name] = s.Address
	}

	for _, h := range r.Hints {
		if serverMap[h.ServerName] == "" {
			return errors.Errorf("No server name found for hint, db:%s, col:%s, server:%s", h.Database, h.Collection, h.ServerName)
		}
	}
	return nil
}

// DetermineServer returns a matching server entry.
func (r *Router) DetermineServer(db, c string) (ret *Server) {
	if s, ok := r.HintMap[db+":"+c]; ok {
		return s
	}
	if s, ok := r.HintMap["db@"+db]; ok {
		return s
	}
	if s, ok := r.HintMap["c@"+c]; ok {
		return s
	}

	return r.Servers[0]
}

func getRouterConfigPath() string {
	if *dbRouterJson != "" {
		return *dbRouterJson
	}
	path := os.Getenv("DB_ROUTER")
	if path != "" {
		return path
	}
	return ""
}

func getDefaultRouter() *Router {
	lock.Lock()
	defer lock.Unlock()
	once.Do(func() {
		var err error
		defaultRouter, err = LoadJsonFileRouter(getRouterConfigPath())
		if err != nil {
			goutils.LogFatal(err)
		}
	})
	return defaultRouter
}

func getClient(db, c string) *DbClient {
	router := getDefaultRouter()
	s := router.DetermineServer(db, c)
	return Dial(s.Address)
}

// Open returns a mgo collection and session.
func Open(db, c string) (*mgo.Collection, *DbSession) {
	col, session := getClient(db, c).Open(db, c)
	return col, session
}

// Open returns a mgo collection and session with long session timeout.
// It's useful for task specific query, which may need dozens of time to
// complete.
// NOTE: It's risky to use such session in server, where may lead to infinite
// no response status.
func OpenWithLongTimeout(db, c string) (*mgo.Collection, *DbSession) {
	router := getDefaultRouter()
	s := router.DetermineServer(db, c)

	col, session := Dial(s.Address).OpenWithLongTimeout(db, c)
	return col, session
}
